# Design

## Context

`workspace_events` (from workspace statistics) already holds one row per
start, refusal and failure, with backend, project and duration, and is
cleaned after `CHATCMS_WORKSPACE_HISTORY_DAYS`. The start's steps and log
live in memory (`progress`) and vanish with the workspace. `builds` keeps
each build's log and exit; `publishes` keeps steps, trouble and detail. The
Hugo program for a project is chosen in `internal/app/hugo.go`
(`hugos.For`), which knows whether it pinned, fell back to the installed one,
or refused.

## Goals / Non-Goals

**Goals:**
- After the fact, see which Hugo a start asked for, which program ran, why,
  and on which backend.
- One page per project where starts, builds and publications sit together.

**Non-Goals:**
- Git level, the assistant, organisation, platform and access logs (bean
  "Logs per area").
- Full-text search, export, or live streaming beyond polling a running start.

## Decisions

- **Extend `workspace_events`, do not add a table.** A start row gains
  `person_id`, `steps` (JSONB), `log` (TEXT, last 400 lines) and `hugo`
  (JSONB: `asked`, `program`, `because`, `printed`). One cleanup, one
  retention. Refusals keep their row without a log.
- **The row is written when the start begins and updated when it ends**, so
  a running start is on the page and can be polled. A server that stops
  mid-start leaves a row that is marked interrupted on start-up.
- **Why a program was chosen comes from the resolver.** `hugos.For` returns
  the program together with a reason: `pinned`, `pinning-off`, `no-version`
  or `not-hugo`; the runtime carries it into the spec, the manager logs it.
- **What the program printed** is the first line of the start's log that
  matches `hugo v<digits>`; for another program, its first line of output.
  Found from the log, so no extra process runs inside the sandbox.
- **The page merges three sources** on the server: starts from
  `workspace_events`, builds and publications from their tables, newest
  first, paged by time (`before=`), filtered by `kind=`. An entry's address
  is `kind:id`.
- **Navigation:** Log under Setup, shown when the project answers that the
  person may see technical detail; History stays as it is.
- **Polling:** an open running entry is read again every second, the list
  every ten seconds while visible.

## Risks / Trade-offs

- Logs can hold paths and hostnames; they are shown only to implementers and
  removed with the history.
- 400 lines may cut the start of a long failure; the end is what usually
  says what went wrong, so the end is kept.
