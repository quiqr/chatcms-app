# Proposal

Beans: [chatcms-2myi](../../../.beans/chatcms-2myi--the-local-backend-lets-a-preview-write-into-the-pe.md)
(the local backend lets a preview write into the person's copy) and
[chatcms-6ure](../../../.beans/chatcms-6ure--live-reload-for-quarto-previews.md)
(live reload for Quarto previews), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

On the local backend Quarto writes its output into the person's copy. A file
there that git does not track made the copy look unsaved: the save
indicator said Saving, the workspace counted as unshared, and other people's
changes stopped arriving. Separately, a Quarto preview did reload by itself
over its websocket, but then went to `/index.html` at the root of the
preview's host, which answers 404, because Quarto keeps a prefix only when
the address has `/p/<word>/` in it.

## What Changes

- **Files nobody tracks are not work.** Whether a copy holds unsaved work,
  and whether it may move forward, ignores untracked files. When a change
  that arrives adds a file a preview had written, the preview's file gives
  way.
- **A rooted preview is addressed as `<prefix>p/site/`.** Quarto then keeps
  the prefix when it reloads; the proxy takes `p/site/` off with the prefix.

## Capabilities

### New Capabilities

### Modified Capabilities

- `runtime/definition`: a rooted preview's address; what a preview writes on
  the local backend.

## Impact

- `internal/gitsync`: status and the fast-forward of a person's copy.
- `internal/workspace`, `internal/preview`: `RootedPath`.
- e2e `38-a-quarto-book`: a Quarto preview reloads by itself on a preview
  origin.

## Assumptions

- **An untracked file is never someone's edit.** Every edit goes through a
  save that commits it under the copy's lock, so between saves the only
  untracked files are output: a preview's, a build tool's.
- **A preview's file that an arrival needs is thrown away**, not kept aside;
  it is output and is written again on the next render.
