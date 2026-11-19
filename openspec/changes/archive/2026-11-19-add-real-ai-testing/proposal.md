# Proposal

Bean: [chatcms-d14p](../../../.beans/chatcms-d14p--real-ai-inference-for-developers-and-an-opt-in-rea.md)
(real AI inference for developers, and an opt-in real e2e run), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

Every assistant test ran against a scripted model, so nothing showed that a
real model reads the page, calls the right tool and changes only what was
asked. Developers had no documented way to run ChatCMS with a real model,
and OpenCode Go, which a developer may already pay for, wants headers the
gateway did not send.

## What Changes

- **The gateway sends a user agent of its own**, and, when
  `CHATCMS_AI_SESSION_HEADER` (the module's `ai.sessionHeader`) names one, a
  stable session per person and project in that header. OpenCode Go asks for
  `x-opencode-session`.
- **The gateway logs each call** with its model and tokens, besides
  recording it.
- **A developer's AI settings live outside the repository**, in
  `~/.config/chatcms/ai.env` (or `CHATCMS_AI_ENV`), read by
  `scripts/dev-be.sh`, which says which model the assistant uses. The dev
  shell provides OpenCode, so a local workspace gets its assistant.
- **An opt-in real e2e run**: `E2E_REAL_AI=1` gives the assistant specs the
  real model from that file, checks what a real model can be relied on to
  do, and prints the calls and tokens each test used. The gate keeps the
  scripted model.

## Capabilities

### New Capabilities

- `acceptance/real-ai`: running the assistant's acceptance tests against a
  real model.

### Modified Capabilities

- `ai/gateway`: the user agent, the session header and the call log.

## Impact

- `internal/aigateway`, `internal/app` (config), NixOS module.
- `scripts/dev-be.sh`, `flake.nix` (the dev shell's OpenCode).
- `test/e2e/harness.mjs`, `fixtures.mjs`, specs 35 and 37.
- `docs/TESTING.md`.

## Assumptions

- **OpenCode Go with `deepseek-v4.1-flash`** is the documented example. Both
  assistant specs at both widths took 11 calls, about 31,000 tokens in and
  600 out: about a cent at the list price Go reported.
- **A real model's words are not checked**, only that the page changed as
  asked, shows as changed, and Undo puts it back.
