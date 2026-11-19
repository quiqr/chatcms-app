# Design

## Context

`EditorShell` draws the form, the preview and an assistant column whose
place `arrange()` works out from the width and whether the assistant is open
(`talking`). `EditorPage` keeps the layout in state and opens the assistant
when a selection arrives (`openAssistant`). `ChatPanel` keeps the stream in
its own state, streamed from the server over SSE, and stays mounted while
hidden so nothing is lost.

## Decisions

### 1. Chat is a view, not a toggle

`EditorLayout` gains `chat`. `arrange(width, chosen)` loses `talking`: the
chat shows when `chosen` is `chat`, as a 400 px column when
`width >= 400 + 420`, otherwise alone. The chat panel stays mounted (hidden)
in the other views, so its stream and its running requests survive a switch.

### 2. The view is remembered

`EditorPage` reads and writes the view through `recall`/`remember` under
`chatcms.editor.view.<person>.<project>`, defaulting to Form + preview as
now. A selection switches to `chat` and is remembered like any other choice.

### 3. The badge comes from the panel

`ChatPanel` reports the number of requests still being worked on through
`onRunning(count)`; `EditorShell` shows it on the Chat + preview segment
unless that view is shown. `Segmented` options gain an optional `count`,
drawn like the Publish button's count.

### 4. Cards and summaries

A person's message keeps the selection it was sent with (`text`, `before`,
`after`, and the page title from the editor). The card draws the last
~60 characters of `before`, the selection marked, and the first ~60 of
`after`. Choosing it asks the editor to mark that place: the preview posts
`highlight` to the frame at once (today it only does so when the frame
reloads). A change summary's Show changes opens the change's page and, when
its request had a selection, marks it the same way.

## Risks / Trade-offs

- **Tests that open the assistant by its button** (e2e 35, 37, 42 and the
  shell tests) change to the view switch.
- **A selection no longer leaves the form open**: somebody editing the form
  and selecting in the preview lands in Chat + preview. That is the design's
  intent (§1: the chip lives in Preview and Chat + preview).
