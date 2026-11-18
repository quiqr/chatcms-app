---
# chatcms-mizc
title: AI gateway on the host
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T20:29:59Z
openspec-link: openspec/changes/archive/2026-11-18-add-ai-gateway
parent: chatcms-z0p7
blocked_by:
    - chatcms-7p90
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-18-add-ai-gateway`.

- `internal/aigateway`: an OpenAI-compatible gateway on a listener of its own
  (port 8091), answering only `POST /v1/chat/completions` and
  `GET /v1/models`, only from this machine and `10.231.0.0/16`. It puts the
  provider's key and the configured model in place of what came in, streams
  answers through, and records every call in `ai_usage`.
- A token per workspace, issued when it starts and revoked when it sleeps,
  fails or is reconciled away; stored hashed. A micro VM gets it and the
  gateway's host address on its kernel command line and keeps them in
  `/run/chatcms/ai.env`.
- Limits per organisation: requests a minute (in memory) and tokens a month
  (from usage), refused with plain reasons, OpenAI-shaped so a client shows
  them as they are. Set and seen by the platform admin in an "AI use" section
  of the organisation page.
- NixOS module `ai.providerUrl`, `ai.model`, `ai.keyFile` (a credential),
  `ai.port`; port 8091 open on the TAP devices in nftables and the NixOS
  firewall, and nowhere else. The VM test checks the 401, both rules and that
  the key is not in the store.
- Decided: own Go gateway rather than LiteLLM; provider and model are
  configuration; budgets in tokens, not money.
