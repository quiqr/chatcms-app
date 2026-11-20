---
# chatcms-s59d
title: A runsc backend
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-20T11:45:17Z
openspec-link: openspec/changes/archive/2026-11-19-add-gvisor-workspaces
parent: chatcms-7l81
---

Workspaces run as gVisor containers with the host Nix store bind-mounted
read-only, the same bare-repo working copy flow as the VM backend, and network
access only to the host.

## Summary of Changes

- Workspaces run in rootless gVisor (`runsc`) sandboxes with the Nix store
  read-only and no network; the preview, the agent and the AI gateway are
  reached through sockets (`chatcms-relay`). Delivered with chatcms-7l81.
