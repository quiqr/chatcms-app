# Tasks

## 1. Keeping starts

- [x] 1.1 Migration: `workspace_events` gains `person_id`, `steps`, `log`,
  `hugo`, and `finished_at`; sqlc queries to begin, finish and read starts.
- [x] 1.2 The Hugo resolver returns why it chose a program; the start logs
  the backend, asked version, program and reason; proven by tests for each
  reason.
- [x] 1.3 The manager writes a start when it begins and finishes it with its
  steps, log (last 400 lines) and what the program printed; an interrupted
  start is marked on start-up; proven by Go and integration tests.

## 2. The log

- [x] 2.1 `GET …/log?kind=&before=` and `GET …/log/{entry}`, technical detail
  only (403 otherwise), merging starts, builds and publications; proven by
  httpapi and integration tests.
- [x] 2.2 The Log page under Setup: filters, entries that open to their
  detail, addresses per entry, copy, live running start, the phone layout;
  strings; proven by Vitest.

## 3. Links

- [x] 3.1 The preview column drops its folded log and offers "Show the log"
  or "See what went wrong" to implementers; the workspace symbol links to
  the log; a failed publication links to its entry; proven by Vitest.
- [x] 3.2 e2e: an implementer follows "Show the log" and sees the Hugo facts
  at both widths; an editor has no Log and is refused.

## 4. Ship

- [x] 4.1 `CHANGELOG.md` under [Unreleased].
- [x] 4.2 `scripts/gate.sh ship` green; close chatcms-qru3.
