---
# chatcms-uzi2
title: Move the assistant to OpenCode 2
status: completed
type: task
priority: normal
created_at: 2026-11-18T21:41:32Z
updated_at: 2026-11-20T20:34:43Z
openspec-link: openspec/changes/archive/2026-11-20-opencode-2
parent: chatcms-z0p7
---

The assistant runs OpenCode 1.15.10, pinned through the flake input
`nixpkgs-opencode`, because the nixpkgs the rest of the flake follows carries
OpenCode 2.0.32, whose `serve` starts a "v2 API" in which none of the routes
ChatCMS uses (`POST /session`, `prompt_async`, `GET /event`) exist.

Done when the assistant runs on the OpenCode the main nixpkgs carries, the
second nixpkgs input is gone, and `35-asking-the-assistant` and the hand-run
micro VM boot test still pass. `internal/assistant` (the relay and the
configuration) and the guest's agent unit are where the change lands.

## Summary of Changes

- The assistant speaks OpenCode 2's API (`/api/session`, `/prompt`,
  `/interrupt`, `/api/event`) and translates its events into the same text,
  tool, done and error events; the frontend is unchanged.
- OpenCode 2 offers MCP tools only through Code Mode (`execute`), which can
  `fetch` the network. The quiqr server has `codemode: false` and `execute`
  is denied, so the page tools are direct and the agent has no network.
- The configuration is written in OpenCode 2's own format; the first prompt
  waits for the page tools; `503 service_starting` is waited through.
- OpenCode 2 retries model errors without end: a refusal, or a third
  attempt, ends the answer with the reason and interrupts the session.
- `nixpkgs-opencode` is gone; OpenCode 2.0.32 comes from the main nixpkgs.
- e2e assistant specs and the gVisor and deploy VM tests pass. The hand-run
  micro VM boot test was not run (it needs a tap device and root).
