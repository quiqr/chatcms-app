# Design

## Context

`version.go` embeds `VERSION`; `internal/version` holds `Version`, filled
from it or by `-X` in `flake.nix`, and `String()` for `--version`. The flake
builds the server once in `mkServer`, used by the `chatcmsd` package and the
`go-build` check; the e2e and NixOS VM checks take `self.packages.chatcmsd`.
The server already asks Hugo its version at start-up (`internal/app/templates.go`).
The platform admin area has a dark shell with three entries, and its API
lives under `/api/administration/`, guarded by `ActionAdministerInstallation`.

## Goals / Non-Goals

**Goals:**

- The identifier of the source reaches the deployed binary without making
  every commit rebuild the checks.
- One answer gives the screen everything it shows, and no part of it can
  fail the whole.

**Non-Goals:**

- Contributor names, the list of open source packages, an update check.

## Decisions

### The source identifier only in the deployed package

`mkServer` takes an optional `source`. The `chatcmsd` package passes
`self.shortRev or self.dirtyShortRev or ""` as
`-X .../internal/version.Source=...`; the `go-build` check builds without it.
The e2e and NixOS VM checks switch from `self.packages.chatcmsd` to the
`go-build` check's server, so a commit that changes only documents keeps
every check cached, as the ship tier relies on.

Alternative: pass the identifier everywhere. Rejected, because the server
derivation would change with every commit, and with it the e2e and VM checks,
which is the twelve-minute wait the tiers just removed.

`internal/version` gains `Source`, `Changed()` (the identifier ends in
`-dirty`), `Short()` (without that suffix) and `Describe()`, which gives
"0.1.0 · a3bab43", "0.1.0 · a3bab43, changed" or "0.1.0 · development
build". `String()` uses `Describe()`, so `--version` and the screen agree.

### The license text embedded

`version.go` in the root package also embeds `LICENSE`, as it does
`VERSION`, so the text shown is the one that shipped with the build.

### Asking what it runs on

`internal/app` builds an `About` source with four questions, each with a
two-second limit and each answering "" when it fails:

- Hugo: the version the server already found at start-up, kept rather than
  asked again.
- git: a new read-only `gitsync.GitVersion(ctx)`, running `git --version`
  through the pinned binary, since git is called only from gitsync.
- PostgreSQL: `SHOW server_version` on the pool.
- Go: `runtime.Version()`.

### The answer

`GET /api/administration/about`, behind `RequireActor` and the policy check
for `ActionAdministerInstallation`, answers:

```json
{ "version": "0.1.0", "source": "a3bab43", "changed": false,
  "describe": "0.1.0 · a3bab43",
  "runsOn": { "hugo": "0.152.2", "git": "2.51.0", "postgres": "16.10", "go": "go1.25.4" },
  "license": { "name": "Apache 2.0", "text": "..." },
  "links": { "source": "https://github.com/quiqr/chatcms-app", "...": "..." } }
```

The links are fixed strings in the server, so the frontend holds no
addresses of its own and a fork can change them in one place.

### The screen

`web/src/routes/AboutScreen.tsx` inside `PlatformAdminPage`, entry "About"
with the `Info` icon after History. Sections: Version, Runs on (a definition
list; "unknown" for an empty value), License (name, and the text in the
existing `Disclosure`), Source (three links with `target="_blank"` and
`rel="noreferrer"`), Credits. Labels avoid git words: "Version", "Runs on",
"Source code".

## Risks / Trade-offs

- [The e2e and VM checks no longer test the exact binary that is deployed]
  → the only difference is one `-X` string; `go-build` and the package share
  everything else, and a unit test covers `Describe()` for every case.
- [A slow `git --version` or database holds the screen] → each question has
  its own two-second limit and gives up to "unknown".

## Migration Plan

None. A server built before this change has no `Source` and says
"development build" until the next deploy from the flake.
