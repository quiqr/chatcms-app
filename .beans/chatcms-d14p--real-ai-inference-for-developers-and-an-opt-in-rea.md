---
# chatcms-d14p
title: Real AI inference for developers, and an opt-in real e2e run
status: in-progress
type: task
priority: normal
created_at: 2026-11-19T08:22:50Z
updated_at: 2026-11-19T08:22:50Z
parent: chatcms-z0p7
---

Every assistant test so far runs against a scripted model. Developers need a documented way to run ChatCMS and the assistant specs against a real model: a local settings file outside the repository, the OpenCode Go endpoint (which wants a session header and a user agent) or any cheap OpenAI-compatible key, and an opt-in e2e run with checks a real model can pass. The scripted model stays the default in the gate.
