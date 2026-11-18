---
# chatcms-4b6f
title: Previews on their own origin
status: todo
type: feature
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T12:51:49Z
parent: chatcms-64h1
blocked_by:
    - chatcms-rmne
---

The proxy serves previews from a separate origin with its own short-lived
access token, so site JavaScript has no access to the app's cookies or API.
The NixOS module gains the second host name.
