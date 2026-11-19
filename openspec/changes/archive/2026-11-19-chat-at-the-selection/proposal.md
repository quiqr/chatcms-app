# Proposal

Bean: [chatcms-7hfz](../../../.beans/chatcms-7hfz--a-small-input-over-the-preview-for-a-selection.md)
(a small input over the preview for a selection), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).
Design: `docs/design/chat-requests/CHAT.md` §1–§4, references 01–03.

## Why

Selecting text in the preview shows a round button that sends the selection
to the chat, switching the editor to Chat + preview. The design asks for the
conversation to start where the person is looking: a chip at the selection,
then a small input there, without leaving the page.

## What Changes

- **A "Chat" chip** at the end of a selection, shown only once the selection
  is finished (after the mouse is released, a pause in keyboard selection, or
  touch handles at rest), never for less than two characters, and gone on
  Esc, when the selection is cleared or scrolls out of view.
- **An input card at the selection** (380 px): "What should change?", a send
  button, quick actions (Make it shorter, Simpler words, Fix spelling) that
  send at once, and a footer saying where it goes. It sits 12 px below the
  selection with an arrow, flips above when there is no room, and stays 12 px
  inside the preview. The selection stays marked while it is open.
- **Sending** adds the request to the chat stream and asks the assistant,
  whichever view is shown; the view does not change.
- **Esc or a click outside** closes the card and keeps what was typed as a
  draft in the chat's message box, with the selection quoted.
- **Ctrl/⌘+Shift+K** with a selection opens the card at once.
- **Who and where**: editors and implementers, in Preview and Chat + preview
  (and the phone's preview); not for viewers, not in Form or Form + preview.
- The overlay draws in its own shadow root with the app's look, and takes
  its words from the app, so it reads the same on any site and in the PDF
  viewer.

## Capabilities

### New Capabilities

### Modified Capabilities

- `assistant/selection`: a selection opens a chip and an input at the
  selection rather than the chat.

## Impact

- `web/src/preview/select.ts`: the chip, the card, timing, placement, the
  marked selection, the messages.
- `web/src/routes/Preview.tsx`: setup with the words and whether chat is on;
  `send` and `draft` messages.
- `web/src/routes/EditorPage.tsx`, `ChatPanel.tsx`: a request from the
  overlay goes into the stream; a draft lands in the message box.
- e2e 37 and 46 select through the chip and the card.

## Assumptions

- The rule "no chip across more than one field" needs field markers, which
  only Hugo will have (chatcms-gf1r); the overlay applies it where markers
  are present and ignores it otherwise.
- Without an assistant on the installation there is no chip: a request
  would have nowhere to go.
- When the selection scrolls out of view the card closes into its chip at the
  preview's edge is simplified to: the card follows the selection, and
  closes (keeping the draft) when the selection leaves the view.
