# Design

## Context

See proposal.md for why. What the code does today:

- `gitsync.Sites` keeps one working copy per site at
  `<sites>/<project>/working-copy`. It exchanges with the remote (`Sync`,
  replay on top, conflicts reported per path) and publishes. Writes to one
  copy are serialised by an in-process lock per directory (`siteLock`).
- Every feature opens that copy by project alone: `openSite(ctx, sites,
  projectID)` in `internal/app/publishing.go`, behind the `Sources`
  interfaces of `content`, `media`, `history`, `sitesync`, `publish` and
  `preview`.
- A save queues a `site-sync` job for the project; the job syncs the shared
  copy with the remote. Settling (`sync/settling`) turns a stopped exchange
  into per-field questions.
- `internal/preview.Servers` runs one Hugo per project, with an idle sweep, a
  cap and TCP readiness, all in memory.

## Goals / Non-Goals

**Goals:**

- One seam, `internal/workspace`, that the micro VM and gVisor epics fill
  with another backend and nothing else.
- Per-person copies built on the exchange and settling code that exists,
  rather than on a second merge path.

**Non-Goals:**

- Running anything but Hugo in a workspace. The agent comes with the chat
  epic, Quarto with the runtime definition.
- A per-organisation workspace limit.
- Catching a person's copy up while they look at it (the merging epic).

## Decisions

### 1. The shared copy is the remote of every person's copy

A person's copy lives at `<sites>/<project>/people/<person>/working-copy`,
cloned from the shared copy with `git clone --local` (objects hardlinked).
Its `origin` is the shared copy's path. The existing `Sync` then does what
the spec asks: fetch, replay the person's versions on top, stop with a
`Conflict` when the same file changed. Sending is a push of `HEAD` to the
shared copy's `main`, which runs with `receive.denyCurrentBranch=updateInstead`
so its checked-out tree follows.

The shared copy is no longer written by people, so its tree stays clean and
`updateInstead` does not refuse. Import, templates, go back on the shared
version and the remote sync still write it, all inside gitsync and under its
lock.

The exchange with the shared copy takes the shared copy's `siteLock` as well
as the person's, always in the order shared then person, so a remote sync and
a person's send cannot interleave. A send that loses a race is refused as not
a fast-forward; the job syncs again and retries, at most three times.

`file://` stays forbidden for remotes people name. The person-copy exchange
uses a config that allows the `file` protocol for this one path, built inside
gitsync and never from input.

*Alternatives:* a bare repo per project with every copy, the shared one
included, as a clone. Cleaner, and the micro VM epic needs one anyway, but it
moves the remote sync, publish and import at the same time as everything
else. Branches `people/<id>` inside the shared copy: no separate directories,
but one checked-out tree cannot be two people's preview.

### 2. gitsync's interface stays in sites and versions

New methods, no branch names or paths leaving the package:

```go
func (s *Sites) OpenFor(ctx context.Context, site, person string) (Repo, error)
func (s *Sites) Share(ctx context.Context, site, person string) (SyncResult, error)
func (s *Sites) RemoveFor(site, person string) error
func (s *Sites) Unshared(ctx context.Context, site, person string) (int, error)
```

`OpenFor` makes the copy on first use and brings it level when it has no
exchange in progress. `Share` is bring level then send. `Repo` is unchanged,
so `content`, `media` and `history` keep their code and only receive another
copy.

### 3. Sources take the person

`Site(ctx, projectID)` becomes `Site(ctx, personID, projectID)` in the
`Sources` of `content`, `media`, `history` and `sitesync`. Every handler
already has the actor; passing it is explicit, where hiding it in the context
would make a forgotten person silently open somebody else's copy. `publish`
keeps opening the shared copy and calls `Share` for the publishing person
first.

Settling: a stopped exchange in a person's copy is settled in that copy by
that person, through the same questions. A stopped exchange in the shared copy
(the remote changed outside ChatCMS) is settled as today.

### 4. Saves queue a share job

`SyncSoon(project)` becomes `ShareSoon(project, person)`, a `site-share` job
with subject `<project>/<person>`. It runs `Share`, then queues the existing
`site-sync` of the shared copy. Remote arrivals reach a person's copy through
`OpenFor`, which brings it level on open.

### 5. The workspace package

```go
type Key struct{ Person, Project uuid.UUID }

type Spec struct {
	Key   Key
	Dir   string
	Under string
}

type Endpoint struct {
	Host string
	Port int
}

type Backend interface {
	Name() string
	Start(ctx context.Context, spec Spec, log io.Writer) (Handle, error)
	Alive(ctx context.Context, handle Handle) (Endpoint, bool)
	Stop(ctx context.Context, handle Handle) error
	Apart() bool
}

type Handle string
```

`Workspaces` (the manager) owns state, limits, the sweep, reconcile and the
startup steps; a backend owns only processes. `Handle` is opaque text the
backend writes and reads back, stored in Postgres, so a VM id or container id
fits later. The local backend's handle is `pid:port:hugo-path`; `Alive`
checks that the process exists, that `/proc/<pid>/cmdline` still names that
Hugo (a pid can be reused), and that the port answers.

The manager hands `preview` an `Endpoint`; `preview` builds the reverse proxy
and keeps the policy check and the live-reload rewrite it has today. A
`fake` backend in `internal/workspace/workspacetest` runs the manager's tests
and proves the seam.

### 6. State in Postgres

```sql
CREATE TABLE workspaces (
    id           UUID PRIMARY KEY,
    person_id    UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    backend      TEXT NOT NULL,
    state        TEXT NOT NULL CHECK (state IN ('starting','warm','sleeping')),
    handle       TEXT,
    reason       TEXT,
    last_used_at TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (person_id, project_id)
);
```

The cap counts `starting` and `warm` rows in one statement inside the
transaction that moves a row to `starting`, so two servers or two requests
cannot both take the last place. `last_used_at` is written at most once a
minute per workspace, to keep a busy preview from writing on every request.

The steps and the log of the latest start are kept in memory, a ring of
200 lines per workspace. They describe a start that is happening now; after a
restart there is nothing to show but the state.

### 7. Lifecycle

- Start: reserve a place, record `starting`, make or open the copy (step
  "fetching site"), backend `Start` (steps "starting workspace", "building
  preview"), wait for the port, record `warm` with the handle.
- Sleep: the sweep, once a minute, stops warm workspaces idle longer than
  `CHATCMS_WORKSPACE_IDLE` and records `sleeping`. The copy stays.
- Cleanup: the same sweep, once an hour, removes sleeping workspaces whose
  person may no longer read the project or whose project is archived, and
  those unused for 30 days with `Unshared` at zero.
- Reconcile at start-up: every `starting` or `warm` row is checked with
  `Alive`; alive means `warm`, anything else is stopped and `sleeping`.
- App stop: the manager stops its loops. A backend says whether its
  workspaces run apart from the app (`Apart() bool`). The local backend's do
  not: Hugo runs in the server's cgroup, so systemd ends it with the server
  anyway, and a test server left behind would leak a Hugo per test. So the
  local backend's workspaces are stopped and recorded as sleeping, and the
  micro VM and gVisor backends will answer true. After a crash, reconcile
  reattaches to a Hugo still running or stops it.

### 8. HTTP and frontend

- `POST /api/orgs/{org}/projects/{project}/workspace` starts the caller's
  workspace and answers at once with its state.
- `GET` of the same path gives the state and, only with
  `ActionSeeTechnicalDetail`, the steps and the log.
- The editor shell posts on opening a project. `Preview.tsx` polls the `GET`
  every 500 ms while starting, shows the indicator, and loads the iframe once
  warm. The unfoldable steps follow DESIGN.md's secondary-detail pattern.

### 9. Configuration

`CHATCMS_WORKSPACES_AT_ONCE` (default 8) and `CHATCMS_WORKSPACE_IDLE`
(default 15m). The old preview names are read when the new ones are absent,
with a warning in the log. The NixOS module gains `workspaces.atOnce` and
`workspaces.idle` and keeps the old options through `mkRenamedOptionModule`.

## Risks / Trade-offs

- [Disk: a checked-out tree per person per project] → hardlinked objects;
  cleanup after 30 days; measured in the integration test on the fixture
  site and recorded in the archive summary.
- [A person's change waiting on an unanswered question never reaches publish]
  → the publish screen lists what has not reached the shared version; the
  plain wording is the merging epic.
- [`updateInstead` refuses when the shared tree is dirty] → nothing outside
  gitsync writes it; a refusal is logged with the paths and the share job
  retries after the remote sync.
- [Changing `Sources` signatures touches many handlers and tests] → one
  mechanical task per package, each with its tests green before the next.
- [Reattaching to a wrong process after pid reuse] → the cmdline check in
  `Alive`.

## Migration Plan

On the first start after upgrade there are no workspace rows and no person
copies. The shared copy is where every site already is; copies are made on
first open. Rollback: the previous release ignores `people/` and the table;
saves that reached the shared version are there, saves still waiting in a
person's copy are not, which the release notes say.

## Open Questions

- Whether the 500 ms poll should become server-sent events when the chat
  epic adds streaming anyway.
