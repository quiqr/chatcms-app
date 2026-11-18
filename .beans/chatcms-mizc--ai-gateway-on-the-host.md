---
# chatcms-mizc
title: AI gateway on the host
status: todo
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T12:51:49Z
parent: chatcms-z0p7
blocked_by:
    - chatcms-7l81
---

## Purpose

Workspaces never hold a real API key. The host runs an OpenAI-compatible
endpoint that holds the key, and each workspace gets a token of its own.

## Scope

- A small Go proxy in the backend, OpenAI-compatible, starting with the
  DeepSeek V4.1 Flash pay-as-you-go API.
- A per-workspace token, revoked when the workspace goes.
- Per-organisation rate and budget limits, next to the limits from
  chatcms-qllv.
- Usage logged in Postgres per organisation, project and user.

## Done when

- A workspace can call the model through the gateway with its token and
  cannot reach the provider directly.
- An organisation over budget is refused with a plain reason.

## Open decisions

- Own Go gateway (default) vs LiteLLM. Flag in the proposal.
- Check the OpenCode Go terms before using it for customers; the PoC uses the
  pay-as-you-go API.

## References

`docs/BRIEFING-M07.md` step 3.
