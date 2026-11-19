# Tasks

## 1. The view

- [x] 1.1 `arrange` with the `chat` view; `Segmented` with a count; proven by
  Vitest.
- [x] 1.2 `EditorShell`: four views, the chat column, no Assistant button on
  the desktop, the badge; proven by Vitest.
- [x] 1.3 `EditorPage`: the remembered view; a selection opens Chat + preview;
  proven by Vitest.

## 2. The chat

- [x] 2.1 `ChatPanel` reports running requests; proven by Vitest.
- [x] 2.2 Request cards with the snippet and the caption; choosing one marks
  the selection in the preview; `Preview` posts a mark at once; proven by
  Vitest.
- [x] 2.3 Change summaries with Show changes and Undo; strings.

## 3. Proof and ship

- [x] 3.1 e2e 35, 37 and 42 through the view switch; e2e 46 for the view, the
  card and the badge at 1440 px; screenshot next to reference 08.
- [x] 3.2 `CHANGELOG.md` under [Unreleased].
- [x] 3.3 `scripts/gate.sh ship` green; close chatcms-xxhd and chatcms-b7az.
