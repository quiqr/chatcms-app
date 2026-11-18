---
# chatcms-7p90
title: Micro VM workspace backend
status: todo
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T12:51:48Z
parent: chatcms-z0p7
blocked_by:
    - chatcms-9iuz
---

## Purpose

The strongest isolation backend: each workspace in its own micro VM, so the
agent in the next epics can run untrusted code. Needs KVM.

## Scope

- microvm.nix on Cloud Hypervisor; host `/nix/store` shared read-only over
  virtio-fs.
- TAP networking with a private IP per VM. The VM reaches only the host (git
  remote, AI gateway). No vsock.
- The host keeps a bare repo per project; the VM clones it, works on the
  user's branch and pushes back to the host. Deploy keys never enter the VM;
  the host publishes.
- The proxy maps workspace to VM IP:port, readiness by TCP dial, live-reload
  websocket proxied.
- Hugo version pinned per project in its quiqr config, realised on the host
  with Very Nix before the VM starts, run from the shared store. Supersedes
  chatcms-90m8.

## Done when

- A preview runs inside a VM with the project's pinned Hugo version.
- The acceptance tests of the first two epics pass on this backend.
- The NixOS VM test covers it (release tier; needs nested virtualisation).

## References

`docs/BRIEFING-M07.md` step 2. A change under `nix/` runs
`scripts/gate.sh release` by hand before shipping.
