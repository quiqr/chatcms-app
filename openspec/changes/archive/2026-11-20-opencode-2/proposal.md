# Proposal

Bean: [chatcms-uzi2](../../../.beans/chatcms-uzi2--move-the-assistant-to-opencode-2.md)
(move the assistant to OpenCode 2), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

The assistant runs OpenCode 1.15.10 from a second nixpkgs input, because
OpenCode 2, which the main nixpkgs carries, serves a new API in which the
routes ChatCMS uses no longer exist. A second nixpkgs is a second set of
packages to keep up to date for one program.

## What Changes

- The assistant speaks OpenCode 2's API: `POST /api/session`,
  `POST /api/session/{id}/prompt`, `POST /api/session/{id}/interrupt` and
  the event stream `GET /api/event`, translating its events into the same
  text, tool, done and error events the chat panel already reads.
- A model that keeps refusing (an error from the gateway) ends the answer
  with what went wrong: OpenCode 2 retries such errors without end, so
  ChatCMS stops the session after a refusal (4xx) at once, or after the
  second retry of anything else.
- OpenCode 2 hands MCP tools to the model only through Code Mode
  (`execute`), which can also `fetch` anything on the network. The
  quiqr server is set to `codemode: false` so its tools are direct tools
  again, and `execute` is denied.
- The configuration is written in OpenCode 2's own format, because the
  `codemode` setting is lost when OpenCode converts the old one.
- The first prompt waits until the quiqr tools are connected, and session
  creation waits while OpenCode answers `503 service_starting`.
- The flake drops `nixpkgs-opencode`; OpenCode comes from the main
  nixpkgs, for the local, gVisor and micro VM backends alike.

## Capabilities

### New Capabilities

### Modified Capabilities

- `assistant/chat`: an answer the model cannot give ends with what went
  wrong.

## Impact

- `internal/assistant/chat.go` and its tests.
- `flake.nix`, `flake.lock`: one nixpkgs.

## Assumptions

- The configuration ChatCMS hands OpenCode (provider, model, MCP server,
  permissions, the build agent's prompt) is read by OpenCode 2 unchanged;
  checked on 20 November 2026 against 2.0.32 with a fake model.
- Event names are from OpenCode 2.0.32 and seen live: `session.text.delta`,
  `session.tool.input.started`, `session.tool.failed`,
  `session.execution.succeeded`, `.interrupted`, `.failed`,
  `session.retry.scheduled`.
