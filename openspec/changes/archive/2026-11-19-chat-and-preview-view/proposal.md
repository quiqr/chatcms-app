# Proposal

Beans: [chatcms-xxhd](../../../.beans/chatcms-xxhd--a-chat-and-preview-view-in-the-editor.md)
(a Chat and preview view in the editor) and
[chatcms-b7az](../../../.beans/chatcms-b7az--a-canvas-request-is-a-card-with-a-thumbnail-of-the.md)
(a canvas request is a card), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).
Design: `docs/design/chat-requests/CHAT.md` §7, reference
`reference/phase-1-chat/screenshots/08-chat-and-preview.png`.

## Why

The assistant is a column that opens beside whatever the editor shows, from
a button of its own. Working only through the assistant (ask, look at the
page, ask again) has no place of its own, and a request made from a
selection shows in the chat as quoted text, without the page around it.

## What Changes

- **A fourth view.** The view switch becomes Form, Form + preview,
  Chat + preview and Preview. Chat + preview puts a 400 px chat column beside
  the preview, without the form. The choice is remembered per person per
  project.
- **No separate Assistant button** on the desktop: the Chat + preview segment
  carries a badge with the number of requests still running.
- **A selection in the preview opens Chat + preview** with the selection
  quoted above the message box, as it now opens the column.
- **A request from the preview is a card**: a snippet of the page with the
  selection marked, the request, and "From the preview · <page>". Choosing
  the card scrolls the preview to the selection and marks it.
- **A change summary** per change: "Changed <page>" with **Show changes**,
  which marks the change in the preview, and **Undo**.
- The phone keeps its current Assistant button until the phone design (§8)
  is built.

## Capabilities

### New Capabilities

### Modified Capabilities

- `interface/scrolling`: the views and where the chat goes.
- `assistant/chat`: request cards and change summaries.
- `assistant/selection`: a selection opens Chat + preview.

## Impact

- `web/src/shells/EditorShell.tsx`, `useEditorRoom.ts`: four views, the chat
  column, the badge.
- `web/src/routes/EditorPage.tsx`: the remembered view, selection opens chat.
- `web/src/routes/ChatPanel.tsx`: cards, summaries, the running count.
- `web/src/routes/Preview.tsx`: a mark sent to a preview already loaded.
- e2e 35, 37, 42 open the chat through the view switch.

## Assumptions

- Where the window has no room for the chat beside the preview (under
  400 + 420 px), Chat + preview shows the chat alone.
- The badge counts requests still being worked on; counting unanswered
  questions waits for §5's question state.
- The snippet is the text around the selection the preview already sends
  (up to 200 characters either side), cut to about 60.
