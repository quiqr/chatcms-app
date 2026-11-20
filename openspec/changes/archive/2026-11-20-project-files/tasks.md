# Tasks

## 1. Server

- [x] 1.1 Files, ReadFile and WriteFile in `internal/content`, with the path
  rules and the sum check; proven by Go tests including `..`, a symbolic
  link and `.git`.
- [x] 1.2 The three routes, technical detail only; proven by httpapi tests.

## 2. The screen

- [x] 2.1 The tree with the ARIA tree pattern; proven by Vitest with the
  keyboard.
- [x] 2.2 The lazily loaded CodeMirror editor with languages, the Go
  template overlay and the theme; packages pinned; `npmDepsHash` updated.
- [x] 2.3 The Files page under Setup: open, edit, save, changed meanwhile,
  not openable, the phone layout; strings; proven by Vitest.

## 3. Through the product

- [x] 3.1 e2e: an implementer opens a template with the keyboard, changes and
  saves it, and the publish count goes up; an editor is refused; both widths.

## 4. Ship

- [x] 4.1 `CHANGELOG.md` under [Unreleased].
- [x] 4.2 `scripts/gate.sh ship` green; close chatcms-nwic.
