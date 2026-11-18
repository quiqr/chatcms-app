---
# chatcms-7p90
title: Micro VM workspace backend
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T18:32:31Z
openspec-link: openspec/changes/archive/2026-11-18-add-microvm-workspaces
parent: chatcms-z0p7
blocked_by:
    - chatcms-a9sj
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-18-add-microvm-workspaces`.

- `internal/workspace/microvm`: each workspace a Cloud Hypervisor micro VM
  booted from one NixOS guest built with microvm.nix (`nix/workspace-guest.nix`,
  flake package and check `workspace-guest`). Two read-only virtio-fs shares,
  the host store and the person's copy; address, Hugo and preview address on
  the kernel command line. A TAP device and a /30 per VM from a pool.
- The NixOS module: `workspaceBackend = "microvm"`, TAP devices `chv<i>` owned
  by the service, an nftables table that lets VMs reach nothing but replies
  from the host, access to `/dev/kvm` and `/dev/net/tun`. The VM test checks
  the devices, their owner and the table. `scripts/dev-taps.sh` for
  development.
- `internal/nixpin`: a project's own Hugo version resolved through Nixhub,
  built with `nix build` behind a GC root, remembered in `nix_packages`.
  Previews and one-shot builds use it; an unknown version is refused naming
  it. On with `pinHugo` (`CHATCMS_NIX_BIN`).
- Measured by hand with the `microvm`-tagged test on an i7-1250U: a cold start
  answers after 4.8 to 6.2 s, a save shows after about 0.6 s (Hugo polls,
  because virtio-fs sends no inotify events), and stopping leaves nothing.
  Faster boots are chatcms-b18f.
- Decided against the briefing: the person's copy is shared into the VM rather
  than cloned and pushed back; the chat epic decides how the VM writes.
- Supersedes chatcms-90m8.
