---
# chatcms-7jbk
title: The Workspace interface, with the preview moved behind it
status: todo
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T12:51:48Z
parent: chatcms-gkdd
---

Introduce `internal/workspace` with a local backend and refactor
`internal/preview` to sit behind it. Proxy, live reload, TCP readiness and the
idle sweep behave as before. Nothing outside the package knows which backend
runs.
