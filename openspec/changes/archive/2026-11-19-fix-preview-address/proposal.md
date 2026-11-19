# Proposal

No bean: a regression found by hand in development, in milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

Since `add-runtime-definition` the pane frames the address the server gives
for a preview. Without a preview origin that address was built from the
app's base URL. Reached by another name (`localhost:5173` in development,
with the base URL on `127.0.0.1:8080`, or any second domain), the frame went
to a host the session cookie does not belong to and answered "Please sign in
first". The selection script likewise posted to the base URL's origin
instead of its own page.

## What Changes

- Without a preview origin, the server gives the preview's path, not a full
  address, so the frame is reached through whatever address the app is.
- Without a preview origin, the selection script and the PDF viewer are told
  no app origin, and answer the page they are on.

## Capabilities

### New Capabilities

### Modified Capabilities

- `preview/origin`: what address a preview has without an origin of its own.

## Impact

- `internal/preview`; `web/src/routes/Preview.tsx`, `web/src/preview/`.
- e2e `41-another-name-for-the-app`.

## Also

- `.chatcms-state/master.key`, the development master key, was tracked by an
  old commit although `.chatcms-state*/` is ignored; every development run
  changed it, and shipping staged it. It is no longer tracked.
- An editor test that needed two edits to land inside the save's timer now
  makes them in one moment, so a loaded machine does not split them.
- The e2e for a cold preview accepts a preview that became ready before the
  steps could be unfolded.

## Assumptions

- With a preview origin nothing changes: its address is absolute, and the
  app's origin is the base URL's.
