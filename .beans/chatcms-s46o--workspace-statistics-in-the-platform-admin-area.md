---
# chatcms-s46o
title: Workspace statistics in the platform admin area
status: completed
type: feature
priority: normal
created_at: 2026-11-18T13:13:56Z
updated_at: 2026-11-19T23:08:04Z
openspec-link: openspec/changes/archive/2026-11-20-workspace-statistics
parent: chatcms-z0p7
blocked_by:
    - chatcms-gkdd
---

## Purpose

The platform admin sees how workspaces are used, so the workspace limit and
the size of the host can be set from numbers rather than guesses.

## Scope

- A Workspaces screen in the platform admin area (chatcms-tr45 shell).
- Now: starting, warm and sleeping, against the limit; per organisation and
  per project; which backend runs them.
- Over time: warm workspaces at once, cold start duration, starts refused at
  the cap, failed starts with their reason.
- Disk held by person copies, per project.
- Read from the `workspaces` table and the manager; no git words, plain
  numbers.

## Done when

- A platform admin sees the current counts and the limit, at 1440 px and
  390 px.
- Nobody else can reach the screen or its API, proven by httpapi tests.

## References

Needs the workspace lifecycle from chatcms-gkdd. Later backends (micro VM,
gVisor) and the AI gateway's usage numbers can join the same screen.

## Summary of Changes

- The workspace manager records starts (cold start time, backend), starts
  refused at the limit and failed starts with their reason, and samples
  starting and warm counts every five minutes.
- History is kept `CHATCMS_WORKSPACE_HISTORY_DAYS` days (90 by default,
  `workspaceHistoryDays` in the NixOS module).
- A Workspaces screen in the platform admin area: now against the limit,
  per organisation and project with disk use, and the past 7 days.
- `GET /api/administration/workspaces`, platform admins only.
