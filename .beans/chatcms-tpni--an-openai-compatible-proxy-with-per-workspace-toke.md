---
# chatcms-tpni
title: An OpenAI-compatible proxy with per-workspace tokens
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T20:29:59Z
parent: chatcms-mizc
---

The backend serves an OpenAI-compatible endpoint, reachable from workspaces
only, that forwards to the provider with the real key. Each workspace gets a
token at start that dies with it. Streaming responses pass through.
