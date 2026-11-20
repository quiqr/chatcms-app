---
# chatcms-uzi2
title: Move the assistant to OpenCode 2
status: in-progress
type: task
priority: normal
created_at: 2026-11-18T21:41:32Z
updated_at: 2026-11-20T19:46:38Z
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
