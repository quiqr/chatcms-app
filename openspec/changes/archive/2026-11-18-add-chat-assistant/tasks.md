# Tasks

## 1. The agent in a workspace

- [x] 1.1 `assistant.OpenCodeConfig`: the JSON and environment; proven by unit
  tests (provider, model, MCP, denied tools, no project config).
- [x] 1.2 `workspace.Endpoint.AgentPort` and `Workspaces.Agent(key)`; proven
  by manager tests.
- [x] 1.3 The local backend starts OpenCode beside Hugo when it has one and a
  token; proven by a local backend test with a stub program and by e2e.
- [x] 1.4 The guest's `chatcms-agent` unit and OpenCode in its closure; the
  microvm backend reports port 4096; proven by the microvm tests and the
  hand-run boot test.

## 2. The tools

- [x] 2.1 The MCP server on the gateway's listener: protocol, token, tools,
  refusals as tool errors; proven by unit tests with fake pages.
- [x] 2.2 The app's pages adapter over `content.Service` and the person's
  versions; proven by `35-asking-the-assistant`, where the real OpenCode
  updates a page through the real MCP server and content service.

## 3. Chat

- [x] 3.1 `assistant.Chat`: sessions, send, event relay, changes fanned out;
  proven by unit tests against a fake OpenCode.
- [x] 3.2 `POST/GET …/chat` and `GET …/chat/events`; proven by httpapi tests.
- [x] 3.3 The chat panel: send, streamed reply, chips with Show and Undo, not
  available; strings in both languages; proven by Vitest.

## 4. Acceptance

- [x] 4.1 `test/e2e/35-asking-the-assistant.spec.mjs` with a scripted model:
  ask, the title changes in the form and the preview, the chip shows, Undo
  puts it back. At 1440 px and 390 px.
- [x] 4.2 `nix/e2e.nix` gives the suite OpenCode.

## 5. Ship

- [x] 5.1 `docs/ARCHITECTURE.md`, `docs/DEPLOYING.md`, `CHANGELOG.md`.
- [x] 5.2 `scripts/gate.sh release` and `scripts/gate.sh ship` green.
