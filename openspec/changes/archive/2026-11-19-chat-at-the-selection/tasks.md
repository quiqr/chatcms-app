# Tasks

## 1. The overlay

- [x] 1.1 Timing: a finished selection by pointer, keyboard and touch; at
  least two characters; Esc and clearing hide the chip; proven by Vitest.
- [x] 1.2 The chip and the card in the shadow root with the app's look;
  words from setup; quick actions; Enter, Shift+Enter, Esc; Ctrl/⌘+Shift+K.
- [x] 1.3 Placement below or above, clamped; follows on scroll; closes when
  the selection leaves; the selection marked while open; proven by Vitest
  on the placement function.
- [x] 1.4 Field markers respected where present.

## 2. The app

- [x] 2.1 `Preview`: setup with words and `on`; `send` and `draft` passed up.
- [x] 2.2 `EditorPage` and `ChatPanel`: a request from the overlay goes into
  the stream without changing the view; a draft fills the message box;
  chat off for viewers and in Form and Form + preview; proven by Vitest.
- [x] 2.3 Strings in `en.ts` and `nl.ts`.

## 3. Proof and ship

- [x] 3.1 e2e 37 and 46 through the chip and the card; e2e 47 for timing,
  flipping, Esc keeping the draft and no chip for a viewer; screenshots next
  to references 01–03.
- [x] 3.2 `CHANGELOG.md` under [Unreleased].
- [x] 3.3 `scripts/gate.sh ship` green; close chatcms-7hfz.
