---
# chatcms-s59d
title: A runsc backend
status: todo
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T12:51:48Z
parent: chatcms-7l81
---

Workspaces run as gVisor containers with the host Nix store bind-mounted
read-only, the same bare-repo working copy flow as the VM backend, and network
access only to the host.
