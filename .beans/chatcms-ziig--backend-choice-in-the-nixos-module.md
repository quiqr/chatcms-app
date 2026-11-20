---
# chatcms-ziig
title: Backend choice in the NixOS module
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-20T11:45:17Z
openspec-link: openspec/changes/archive/2026-11-19-add-gvisor-workspaces
parent: chatcms-7l81
blocked_by:
    - chatcms-s59d
---

The NixOS module gains an option naming the workspace backend (local, gvisor,
microvm) and installs what it needs. The NixOS VM test covers gvisor.

## Summary of Changes

- `services.chatcms.workspaceBackend` names local, microvm or gvisor and
  installs what it needs; the `nixos-vm-gvisor` check covers gvisor.
  Delivered with chatcms-7l81.
