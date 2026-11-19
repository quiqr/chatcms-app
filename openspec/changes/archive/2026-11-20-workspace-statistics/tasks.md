# Tasks

## 1. History

- [x] 1.1 Migration for `workspace_events` and `workspace_samples`; sqlc
  queries; the Postgres history; proven by integration tests.
- [x] 1.2 The manager records started, refused and failed, samples every five
  minutes and forgets after the configured days; proven by Go tests.
- [x] 1.3 `CHATCMS_WORKSPACE_HISTORY_DAYS` (default 90) and
  `workspaceHistoryDays` in the NixOS module.

## 2. Statistics

- [x] 2.1 The statistics: now, per organisation and project, the past 7 days,
  disk per project; proven by integration tests.
- [x] 2.2 `GET /api/administration/workspaces`, admins only; proven by
  httpapi tests.

## 3. The screen

- [x] 3.1 The Workspaces screen and its navigation item; strings; proven by
  Vitest.
- [x] 3.2 e2e: a platform admin sees the screen at both widths; an editor is
  refused.

## 4. Ship

- [x] 4.1 `CHANGELOG.md` under [Unreleased].
- [x] 4.2 `scripts/gate.sh release` (nix/ changes) green; close chatcms-s46o.
