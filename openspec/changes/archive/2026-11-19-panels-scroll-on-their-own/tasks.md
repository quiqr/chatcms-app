# Tasks

## 1. The frame

- [x] 1.1 Desktop shells as a window-tall frame: the top bar stays, panels
  scroll on their own; proven by Vitest on the shells and e2e assertions
  that the header stays at the top after scrolling a long list.
- [x] 1.2 `SitePanel` folds to its rail below the wide width and opens as a
  drawer, leaving the person's fold alone; proven by Vitest.

## 2. The editor

- [x] 2.1 The form, preview and assistant slots take the window's height,
  not the form's; proven by an e2e on a page with a long form: the preview's
  bottom and the assistant's message box are inside the viewport.
- [x] 2.2 `useEditorRoom`: side by side only where both fit; the switch at
  every desktop width; the form first; proven by Vitest and an e2e at a
  medium width (1024 px).
- [x] 2.3 The assistant in the form's place where it has no room, beside the
  preview; the form back when it closes; proven by Vitest and e2e 35 and 37
  at 1440 px.

## 3. Ship

- [x] 3.1 `CHANGELOG.md` under [Unreleased].
- [x] 3.2 `scripts/gate.sh ship` green; close chatcms-x9uq and chatcms-y9bd.
