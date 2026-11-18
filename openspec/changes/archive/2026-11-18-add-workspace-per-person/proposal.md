# Proposal

Bean: [chatcms-gkdd](../../../.beans/chatcms-gkdd--workspace-per-user-per-project-local-backend.md)
(workspace per user per project, local backend), the first epic of milestone
07 ([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

A project has one working copy and one Hugo server, shared by everyone who
edits it. A chat agent, a micro VM and select-to-edit all need the opposite:
a place of one's own, per person, that can run untrusted work and be moved
somewhere else without the rest of the app noticing. This change adds that
place, still running on this machine, and the seam later epics swap the
machine behind.

## What Changes

- **A workspace per person per project.** The app asks for workspace(person,
  project) and gets an address and a state. `internal/workspace` owns the
  lifecycle; a backend owns where it runs. The first backend, `local`, runs
  Hugo as a process on this machine, as the preview does today.
- **A working copy per person.** Each person gets their own copy of the site,
  cloned from the shared working copy. Saving writes there, so two people can
  preview different unfinished changes at the same time.
- **Saves reach the shared version on their own.** After a save, the person's
  copy exchanges with the shared copy through the engine that already
  exchanges with the remote: replay on top, then send. When two people
  changed the same field, the existing settling questions are asked. The
  shared copy keeps talking to the remote and publishing as before.
- **Lifecycle in Postgres.** A workspace is starting, warm or sleeping. On
  start-up the server reattaches to workspaces still running or stops them.
  The local backend's workspaces stop with the app; later backends' keep
  running across a restart.
  A sleeping workspace whose person can no longer read the project, or that
  has gone unused for 30 days with nothing unshared, is cleaned up.
- **Workspace limits replace the preview cap.** `CHATCMS_WORKSPACES_AT_ONCE`
  and `CHATCMS_WORKSPACE_IDLE` replace `CHATCMS_PREVIEWS_AT_ONCE` and
  `CHATCMS_PREVIEW_IDLE`; the old names are still read for one release.
- **Eager start and a preparing indicator.** Opening a project starts its
  workspace. Content creators see "preparing preview"; implementers and
  owners can unfold the steps and their log.
- `internal/preview` keeps the proxy, the policy check and live reload, and
  asks `internal/workspace` for where to send a request.

## Capabilities

### New Capabilities

- `workspace/lifecycle`: what a workspace is, who it belongs to, when it
  starts, sleeps and is cleaned up, how many run, and what happens to them
  when the server restarts.
- `workspace/startup`: what a person sees while their workspace starts, and
  who may see the steps and the log.
- `sync/personal-copies`: each person's own copy of a site, how saves made
  there reach the shared version, and what the shared version still does
  alone.

### Modified Capabilities

- `preview/lifetime`: one preview per person per site instead of per site;
  the idle stop becomes a workspace going to sleep; the cap becomes the
  workspace limit.
- `preview/serving`: a preview shows the asking person's own copy.

## Impact

- **Seams.** `gitsync` gains a person's copy in its own vocabulary
  (`Sites.OpenFor(ctx, site, person)` and an exchange with the shared copy);
  no branch name or path leaves the package. The `Sources` interfaces of
  `content`, `media`, `history` and `sitesync` take the person as well as the
  project. `runner` and `policy` are untouched; builds stay one-shot jobs.
- New `internal/workspace` with the `local` backend; `internal/preview`
  shrinks to proxying.
- Migration `00018_workspaces.sql`, sqlc queries, config, the NixOS module
  options.
- Frontend: the editor shell starts the workspace; the preview column shows
  the indicator and, for implementers, the steps.
- Acceptance: keeps M01 acceptance test 5 (a saved change shows in the
  preview) and adds "two people preview the same project at once with
  different changes", the definition of done of this epic.

## Assumptions

- Viewers get a copy of their own too. It is never written to, and one rule
  for everyone is simpler than a read-only path; objects are hardlinked, so
  the cost is a checked-out tree.
- The workspace limit is per installation. A per-organisation limit waits for
  a customer who needs it.
- Most of the merging epic (chatcms-9iuz) lands here, because reusing the
  exchange and settling engine is cheaper than building a temporary path.
  That epic keeps catching up when others save, publish wording, the plain
  conflict wording and soft locking.
- Cleanup after 30 days unused is a constant, not a setting.
