# Design

## Context

The `workspaces` table holds the current state of each person's workspace
per project; the manager (`internal/workspace`) reserves a place within the
limit in `Reserve`, starts in `start`, and sweeps every minute and cleans up
every hour. Nothing remembers what happened.

## Decisions

### 1. Two tables

`workspace_events (id, at, kind started|refused|failed, backend, project_id,
millis, reason)` and `workspace_samples (at, starting, warm)`. Events are
few (one per start); samples are one row per five minutes, about 26 000 in
90 days.

### 2. The manager records through an optional History

`Settings.History` (an interface: `Record`, `Sample`, `Forget`) keeps tests
and stores without history working. `begin` records `refused` on
`ErrTooMany`; `start` records `started` with the time from reserving to warm,
or `failed` with `reasonFor(err)` and the first line of the error. The run
loop samples every five minutes from the live workspaces, and the hourly
cleanup forgets history older than `HistoryDays`.

### 3. Statistics come from SQL

Counts now by state, per organisation and per project join `workspaces`,
`projects` and `organisations`. The past 7 days: `count`, `percentile_cont`
for median and p90, refused, failed grouped by reason, `max(warm)` per day.
Disk is the size of `sites/<project>/people/*`, walked on request with a
deadline per project.

### 4. One endpoint, one screen

`GET /api/administration/workspaces` checks `MayAdminister`, like
`/administration/about`. The screen lives at `/administration/workspaces`
in the platform admin shell, with a navigation item; tables become lists at
390 px.

## Risks / Trade-offs

- **Walking copies** on request costs time on large installations; bounded
  per project, and only on this screen.
- **Samples miss short peaks** between five-minute marks; the cap refusals
  show those.
