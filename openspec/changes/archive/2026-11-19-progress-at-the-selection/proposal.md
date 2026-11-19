# Proposal

Bean: [chatcms-6s8v](../../../.beans/chatcms-6s8v--progress-and-undo-next-to-the-selection.md)
(progress and Undo next to the selection), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).
Design: `docs/design/chat-requests/CHAT.md` §5, references 04 to 07.

## Why

A request sent from the selection disappears into the chat: in Preview the
person sees nothing until the page reloads, and to undo it they have to open
the chat. The design keeps the state where the request was made.

## What Changes

- **A pill at the selection** replaces the input once a request is sent:
  - **Working**: "Working on it…" and **Stop**, which stops the assistant.
  - **Done**: "Changed", **Undo** and **Show in chat**. The changed text is
    tinted green for a moment. The pill stays at least 8 seconds and until
    the person touches the preview again; then Undo lives in the chat only.
  - **Question**: "The assistant has a question" and **Answer in chat**, which
    opens Chat + preview with the message box ready.
- **Several at once**: every request keeps its own pill; the Chat + preview
  segment counts running requests and unanswered questions.
- **Undo from the pill** puts back every change that request made, in one
  step, and the chat says "Undone".
- **The server can stop** a running answer (`POST …/chat/stop`), aborting the
  assistant's session.

## Capabilities

### New Capabilities

### Modified Capabilities

- `assistant/selection`: a sent request shows its progress at the selection.
- `assistant/chat`: a running answer can be stopped; questions count.

## Impact

- `internal/assistant`, `internal/httpapi`: stopping a running answer.
- `web/src/preview/select.ts`: pills, their states and actions, the green tint.
- `web/src/routes/Preview.tsx`, `EditorPage.tsx`, `ChatPanel.tsx`: request
  states flow from the chat to the preview and back; pills survive a reload
  of the preview.

## Assumptions

- **A question** is an answer that ends with a question mark and changed
  nothing; the assistant has no separate signal for it. It counts as
  unanswered until the person sends the next message.
- **One at a time**: the assistant answers one request at a time; requests
  sent meanwhile wait and show Working.
- **Stop** stops the request being answered; a waiting request is taken out
  of the queue.
- **Pills that would overlap** stack under each other, and a pill whose
  context changed (the text before it was the one just edited) finds its
  place by the selected text itself.
