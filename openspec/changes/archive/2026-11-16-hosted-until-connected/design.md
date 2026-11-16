# Design

## Context

`gitsync.Sites.Create` makes a working copy from files and sets `origin` to
the remote it is given; an empty remote leaves it without one. The template
import calls `Publish` right after, which sends the first version to
`origin`. `Keys.PathForSite` gives no key path when a project has no key, so
a remote that is a path on disk works without one. Nothing yet changes a
project's remote after the import, although `tenancy/projects` already
promises it.

## Goals / Non-Goals

**Goals:**

- A kept site behaves like any other in every screen except publishing.
- Connecting later is one action with its own refusals, and leaves no half
  state behind.

**Non-Goals:**

- Joining a kept site's history with a repository that already holds work.
- Serving kept repositories to anything outside ChatCMS.

## Decisions

### A kept site is a bare repository in the state directory, set as origin

`Sites.Host(ctx, id)` makes `<state>/hosted/<id>.git` with
`git init --bare --initial-branch=main` and returns its path. The import uses
that path as the remote, so `Create`, `Publish`, sync and history need no
special case. A `hosted` flag on the project says which sites are kept; the
path alone is not used to decide it, so a moved state directory cannot
change what a project is.

Alternative: no remote at all until connected. Every place that talks to
`origin` would then need a branch for "there is none", and the first version
of a template would have nowhere to go.

### Connecting moves the remote in one gitsync operation

`Repo.MoveTo(ctx, remote)` sets `origin` to the new address and pushes the
main line there. The service checks first, in this order: the person may
change settings, the site is kept, the key reaches the address
(`Reachable`), the address is empty (`RemoteIsEmpty`). Only after the push
succeeds does it record the address, clear `hosted`, and remove the kept
repository. If the push fails, `origin` is set back, so a failed connect
leaves the project as it was.

Alternative: clone the kept repository into a new working copy. That throws
away a working copy that may hold unsent work, and changes nothing a user
would notice.

### Publishing checks the flag before anything runs

The publish service reads `hosted` and refuses with the key
`refused.site-not-connected` before a publish record is made, so the
history of publishes stays a list of real attempts.

### The screens say "kept in ChatCMS", never a path

The step's git address becomes optional, with a hint saying an empty one
keeps the site in ChatCMS until it is connected in settings. With no address,
Continue copies the template in before opening the setup page, since there is
no key to wait for; if the copy fails, the setup page offers it again. The
setup page skips the key while hosted. Project settings
show "Kept in ChatCMS, not connected yet" with the key and an address field
to connect. The path of the kept repository is never shown.

### A project with nothing brought in says so

Every screen that opens a working copy answers a project with none with the
refusal `refused.nothing-brought-in-yet`, not a server error.

## Risks / Trade-offs

- [Kept repositories are lost with the state directory] → `docs/DEPLOYING.md`
  lists `<state>/hosted/` under what to back up, and the settings screen says
  the site is not connected.
- [Somebody connects the wrong repository] → only an empty repository is
  accepted, so nothing of theirs can be overwritten.
