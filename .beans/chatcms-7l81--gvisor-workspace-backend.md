---
# chatcms-7l81
title: gVisor workspace backend
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-20T02:00:00Z
openspec-link: openspec/changes/archive/2026-11-19-add-gvisor-workspaces
parent: chatcms-z0p7
blocked_by:
    - chatcms-7p90
---

## Purpose

The candidate first production backend: gVisor (runsc) containers run on any
VPS without KVM, cheap and easy to scale.

## Scope

- A runsc backend with the Nix store as a read-only bind mount and ordinary
  container networking, limited to the host like the VM backend.
- The installation chooses its backend in the NixOS module.
- A measurement of file I/O on a large Hugo site, recorded in the proposal,
  since gVisor is slower there and isolates less than a VM.

## Done when

- The acceptance tests of the first three epics pass on gVisor.
- Switching backend is a configuration change, with no code change outside
  `internal/workspace`.

## References

`docs/BRIEFING-M07.md`, addendum "Workspace backends".

## Summary of Changes

Workspaces can run in gVisor sandboxes (`workspaceBackend = "gvisor"`),
rootless and without network: the sandbox sees the Nix store read-only, the
person's copy, a private /tmp and its socket directory; the preview, the
agent and the AI gateway are reached only through Unix sockets joined to
ports by `chatcms-relay`. OpenCode runs through glibc's loader
(`opencode-sandboxed`), as gVisor's ELF loader refuses its binary.

- Proven by a boot test with the real runsc (preview through its socket,
  internet and host blocked, nothing left after Stop), an agent boot test,
  a NixOS VM test under the service's own fences, and an acceptance run on
  gVisor (03, 04, 05, 33, 35, 36, 37, 46, 48).
- Measured on 3000 pages: a full build 0.4 s locally, 2.0 s in gVisor;
  a sandbox starts in about 0.5 s.
- `CHATCMS_DEV_BACKEND=gvisor scripts/dev-be.sh` runs it in development.
