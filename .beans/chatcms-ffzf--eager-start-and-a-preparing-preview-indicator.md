---
# chatcms-ffzf
title: Eager start and a preparing-preview indicator
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T14:44:32Z
parent: chatcms-gkdd
blocked_by:
    - chatcms-cggu
---

Opening a project starts its workspace, so the preview is usually ready when
asked for. Content creators see one "preparing preview" indicator.
Implementers can expand it to the startup steps (starting workspace, fetching
site, building preview) and their logs. Target: 1-3 s cold, instant warm.
