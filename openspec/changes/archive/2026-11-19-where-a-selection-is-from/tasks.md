# Tasks

## 1. The matcher

- [x] 1.1 `internal/anchor`: plain text, pages from the model, fields and
  bodies, data files by line, the order and the context rule; proven by Go
  tests.

## 2. Using it

- [x] 2.1 The assistant's prompt names the field, shared text or the design;
  proven by Go tests.
- [x] 2.2 `POST …/chat/where`; proven by an httpapi test.
- [x] 2.3 The placer in the app reads the person's copy; proven by a Go test.

## 3. Ship

- [x] 3.1 e2e 37 still sees the page key in what the assistant was asked.
- [x] 3.2 `CHANGELOG.md` under [Unreleased].
- [x] 3.3 `scripts/gate.sh ship` green; chatcms-gf1r stays open for tier 1.
