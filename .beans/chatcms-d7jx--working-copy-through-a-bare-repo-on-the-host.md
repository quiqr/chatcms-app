---
# chatcms-d7jx
title: Working copy through a bare repo on the host
status: todo
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T12:51:48Z
parent: chatcms-7p90
blocked_by:
    - chatcms-zuer
---

The host keeps a bare repo per project. The VM clones it, works on the user's
branch and pushes back to the host. Deploy keys stay on the host, and
publishing to the outside world happens there.
