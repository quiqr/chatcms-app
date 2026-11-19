# Tasks

## 1. Server

- [x] 1.1 `Report` carries the backend and when the start began; proven by
  Go tests.
- [x] 1.2 The standing answers `since` for everybody and `backend` for
  technical detail only; proven by tests.
- [x] 1.3 `POST …/workspace/restart`: sleep then open, technical detail only;
  proven by httpapi and preview tests.

## 2. The symbol

- [x] 2.1 `WorkspaceIndicator` beside `SyncIndicator` in the editor, project
  and phone shells, with its popover; strings; proven by Vitest.
- [x] 2.2 e2e: an implementer sees the backend and starts again; an editor
  sees plain words only; at both widths.

## 3. Ship

- [x] 3.1 `CHANGELOG.md` under [Unreleased].
- [x] 3.2 `scripts/gate.sh ship` green; close chatcms-e6lo.
