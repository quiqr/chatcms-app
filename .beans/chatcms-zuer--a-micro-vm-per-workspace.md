---
# chatcms-zuer
title: A micro VM per workspace
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T18:32:31Z
parent: chatcms-7p90
---

microvm.nix on Cloud Hypervisor, host store shared read-only over virtio-fs,
a TAP device and private IP per VM, firewall rules so the VM reaches only the
host. Started, stopped and reconciled through the Workspace lifecycle.
