---
# chatcms-cggu
title: Workspace lifecycle in Postgres, reconciled on start-up
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T14:44:32Z
parent: chatcms-gkdd
blocked_by:
    - chatcms-8g3i
---

Workspace state (starting, warm, sleeping, gone) lives in Postgres. On start-up
the server reattaches to workspaces still running or cleans them up. Workspace
limits replace `PreviewsAtOnce`, per installation and within the existing
per-organisation limits.
