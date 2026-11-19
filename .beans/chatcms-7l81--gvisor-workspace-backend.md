---
# chatcms-7l81
title: gVisor workspace backend
status: in-progress
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T12:51:48Z
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
