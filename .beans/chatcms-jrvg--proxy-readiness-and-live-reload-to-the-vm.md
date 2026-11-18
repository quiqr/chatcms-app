---
# chatcms-jrvg
title: Proxy, readiness and live reload to the VM
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T18:32:31Z
parent: chatcms-7p90
blocked_by:
    - chatcms-d7jx
---

The preview proxy maps the workspace to the VM's IP and port, waits for it by
TCP dial and proxies the live-reload websocket. Behaviour matches the local
backend.
