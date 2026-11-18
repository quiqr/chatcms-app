---
# chatcms-b18f
title: Micro VM workspaces boot in under three seconds
status: todo
type: feature
priority: normal
created_at: 2026-11-18T17:53:57Z
updated_at: 2026-11-18T17:53:57Z
parent: chatcms-z0p7
blocked_by:
    - chatcms-7p90
---

A cold micro VM workspace answers after about 5.3 s on the development
machine (i7-1250U): about 3.6 s until the guest's second-stage systemd,
kernel and initrd, and the rest until Hugo answers. The briefing aims at
1-3 s cold.

Candidates, measured one at a time:

- a kernel with virtio built in and no modules to load from the initrd;
- booting without the NixOS initrd stage, root on virtio-fs;
- restoring a booted guest from a Cloud Hypervisor snapshot and handing it
  its address, site and Hugo afterwards.

Done when a cold start of a typical site answers within 3 s on the same
machine, measured the same way as in
`openspec/changes/archive/*-add-microvm-workspaces/design.md`.
