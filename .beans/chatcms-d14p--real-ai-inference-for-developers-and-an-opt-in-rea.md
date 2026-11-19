---
# chatcms-d14p
title: Real AI inference for developers, and an opt-in real e2e run
status: completed
type: task
priority: normal
created_at: 2026-11-19T08:22:50Z
updated_at: 2026-11-19T09:27:57Z
openspec-link: openspec/changes/archive/2026-11-19-add-real-ai-testing
parent: chatcms-z0p7
---

Every assistant test so far runs against a scripted model. Developers need a documented way to run ChatCMS and the assistant specs against a real model: a local settings file outside the repository, the OpenCode Go endpoint (which wants a session header and a user agent) or any cheap OpenAI-compatible key, and an opt-in e2e run with checks a real model can pass. The scripted model stays the default in the gate.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-19-add-real-ai-testing`.

- The gateway sends a user agent of its own, a stable session per person and
  project in `CHATCMS_AI_SESSION_HEADER` when named (`ai.sessionHeader`;
  OpenCode Go asks for `x-opencode-session`), and logs each call's tokens.
- `~/.config/chatcms/ai.env` (or `CHATCMS_AI_ENV`) holds a developer's AI
  settings; `scripts/dev-be.sh` reads its `CHATCMS_AI_*` lines and says
  which model the assistant uses. The dev shell provides OpenCode.
- `E2E_REAL_AI=1` runs specs 35 and 37 against the real model, checking
  the change and Undo rather than the words, and prints each test's use.
  With OpenCode Go and `deepseek-v4.1-flash`, both at both widths: 11
  calls, about 31,000 tokens in, 600 out, about a cent.
