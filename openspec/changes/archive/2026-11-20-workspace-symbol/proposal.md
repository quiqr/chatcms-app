# Proposal

Bean: [chatcms-e6lo](../../../.beans/chatcms-e6lo--workspace-or-vm-symbol-with-extra-info-when-clicke.md)
(a workspace symbol with more when clicked), under
[chatcms-iqog](../../../.beans/chatcms-iqog--ui-cleanups.md) (UI cleanups).

## Why

Whether a person's own preview runs, is starting, or sleeps is only visible
while the preview column is open. Nobody can tell whether the assistant is
there, and an implementer has no way to start a stuck workspace again short
of waiting for it to sleep.

## What Changes

- **A workspace symbol** next to the save status, in the editor, the project
  and the phone shells, coloured by state: starting, ready, sleeping, did not
  start.
- **Clicking it opens a popover** that says, in plain words:
  - whether the preview is ready, being prepared, asleep or did not start,
    and since when;
  - whether the assistant can be asked.
- **For implementers** the popover also says on what the workspace runs
  (local, micro VM, gVisor), the start steps and the log, and has a
  **Start again** button.
- The workspace standing gains `since` (when this start began) for
  everybody, and `backend` for those who may see technical detail.
- **`POST …/workspace/restart`**: puts the person's workspace to sleep and
  opens it again; only for somebody who may see technical detail.

## Capabilities

### New Capabilities

- `workspace/symbol`: the workspace symbol, its popover and starting again.

### Modified Capabilities

## Impact

- `internal/workspace`: `Report` carries the backend and the start time;
  a restart.
- `internal/preview`, `internal/httpapi`: the standing's new fields and the
  restart endpoint.
- `web/src/shells`: the symbol beside `SyncIndicator`; strings.

## Assumptions

- Agreed on 19 November 2026: a symbol beside "All changes saved", with a
  popover; plain words for editors, more for implementers.
- The symbol polls the standing every 10 seconds while the page is visible,
  and every second while starting; the preview column keeps its own polling.
- "Since" is when the current start began; a sleeping workspace has none.
- Editors see no backend name and no restart: those are technical detail.
