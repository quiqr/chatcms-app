# Proposal

Bean: [chatcms-s46o](../../../.beans/chatcms-s46o--workspace-statistics-in-the-platform-admin-area.md)
(workspace statistics in the platform admin area), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

How many workspaces may run at once, and how large the host must be, are
set from guesses. The platform admin should see how workspaces are used:
now, and over the past days.

## What Changes

- **History**: the workspace manager records every start (with its cold
  start time and backend), every start refused because the limit was
  reached, and every failed start with its reason; and every five minutes
  how many workspaces are starting and warm. History older than
  `CHATCMS_WORKSPACE_HISTORY_DAYS` (default 90, `workspaceHistoryDays` in the
  NixOS module) is removed by the hourly cleanup.
- **A Workspaces screen** in the platform admin area:
  - now: starting, warm and sleeping against the limit, and the backend;
    the same per organisation and per project;
  - the past 7 days: starts, median and slow (90th percentile) cold start
    time, starts refused at the limit, failed starts by reason, and the most
    warm at once per day;
  - disk held by the people's copies, per project.
- **`GET /api/administration/workspaces`**, for platform admins only.

## Capabilities

### New Capabilities

- `workspace/statistics`: recording workspace history and showing it to the
  platform admin.

### Modified Capabilities

## Impact

- `db/migrations/00022_workspace_history.sql`, `db/queries/workspaces.sql`.
- `internal/workspace`: recording in the manager; the history store;
  statistics.
- `internal/httpapi`: the endpoint; `internal/app`: wiring and the setting.
- `web/src/routes`: the screen; the admin navigation.
- `nix/nixos-module.nix`: `workspaceHistoryDays`.

## Assumptions

- History is kept 90 days by default, configurable (agreed 19 November
  2026); the screen shows the last 7.
- Disk is measured on request by walking the copies, bounded in time; a
  project whose walk does not finish says so instead of a number.
- Plain numbers, no git words; the screen is for the platform admin only.
