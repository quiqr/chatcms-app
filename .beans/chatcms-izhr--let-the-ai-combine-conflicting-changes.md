---
# chatcms-izhr
title: Let the AI combine conflicting changes
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-19T10:07:29Z
openspec-link: openspec/changes/archive/2026-11-19-add-ai-combine
parent: chatcms-9iuz
blocked_by:
    - chatcms-7j75
---

The plain conflict choice gains a third option: the agent proposes a combined
version, the person accepts or goes back to keep mine or keep theirs.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-19-add-ai-combine`.

- A third choice on each field question, "Let the AI combine them", where
  the assistant is available. The combined text shows in a box the person
  can change; "Use the combination" keeps it, and the two versions stay on
  offer. It is written only when the answers are saved.
- The host asks the model once through `aigateway.Gateway.Ask`, as the
  person, inside their organisation's limits, recorded as their use.
- e2e `39-letting-the-ai-combine`, scripted at both widths and against
  OpenCode Go (`deepseek-v4.1-flash`): "Korenaar Bakery" and "Korenaar on
  the Oude Gracht" became "Korenaar Bakery on the Oude Gracht", one call of
  about 320 tokens.
- Fixed on the way: a person's copy that held history the shared version
  had rewritten never moved forward again; it now follows when it holds
  nothing of its own (`refs/chatcms/taken`).
