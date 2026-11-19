# Design

## Context

The overlay (`select.ts`) sends `send` and `draft`; `ChatPanel` keeps the
stream and answers one request at a time, queueing the rest; changes arrive
as `change` events and Undo calls `goBackTo` with the version before. The
preview reloads after every change, which reinstalls the overlay.

## Decisions

### 1. The chat owns request states; the overlay draws them

Each request from the overlay carries an id. `ChatPanel` reports the state
of every request with an anchor (`working`, `done`, `question`, gone) and
its changes through `onRequests`. `EditorPage` hands the list to `Preview`,
which posts `pills {pills: [{id, state, before, after, text}]}` to the frame
whenever the list changes and again on every `ready`, so a reload loses
nothing. The overlay places each pill with `findBetween` on its anchor,
refreshed on scroll.

### 2. Actions go back as messages

The overlay posts `undo {id}`, `stop {id}` and `open-chat {id}`. Undo asks
the chat to go back to the version before the request's first change, which
undoes all its changes in one step, and marks them undone. Stop calls
`POST …/chat/stop` for the running request, or drops a waiting one. Open
chat switches to Chat + preview and focuses the message box.

### 3. How long Done stays

The overlay notes when a pill turned done. It removes a done pill on the
first pointer or key event in the preview after 8 seconds. The app forgets
it then (`seen {id}`), so it is not drawn again after a reload.

### 4. The green tint

The `chatcms-changed` highlight becomes the success tint
(`rgba(22,163,74,.18)`), drawn over the text between the request's context
once its change arrives, as before.

### 5. Stopping on the server

`Chat.Stop` posts to OpenCode's `/session/{id}/abort` for the person's
session, when there is one. The handler is `POST …/chat/stop`, for a signed-in
person who may use the chat, like asking.

## Risks / Trade-offs

- **Questions are guessed** from a trailing question mark on an answer that
  changed nothing; a wrong guess only shows a different pill.
- **Anchors after a change**: the text between the context is the new text,
  which is what the pill should follow; when the context itself changed,
  `findBetween` fails and the pill is not drawn, while the chat keeps it.
