# Proposal

Bean: [chatcms-7l81](../../../.beans/chatcms-7l81--gvisor-workspace-backend.md)
(gVisor workspace backend), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).
Briefing: `docs/BRIEFING-M07.md`, addendum "Workspace backends".

## Why

Workspaces run as local processes (no isolation) or in micro VMs (strong
isolation, but KVM is needed, which most VPSes do not offer). gVisor
(`runsc`) isolates a workspace in a user-space kernel on any Linux host. It
is the candidate first production backend: cheap, and no special hardware.

## What Changes

- **A `gvisor` workspace backend**: each workspace runs in its own `runsc`
  sandbox, rootless, built from an OCI bundle: an empty read-only root, the
  Nix store read-only, the person's copy of the site at `/site`, a private
  `/tmp`, and one socket directory.
- **No network in the sandbox.** The preview and the agent are reached, and
  the AI gateway reached from inside, through three Unix sockets that runsc
  passes through (`--host-uds`); a small relay (`chatcms-relay`) joins each
  socket to a TCP port inside the sandbox, and the host joins the preview and
  agent sockets to local ports for the proxy. Nothing else leaves it.
- **The same work inside as the micro VM**: the preview command, its PATH
  and environment from the runtime definition, and OpenCode as the agent
  with the gateway's token.
- **The installation chooses**: `services.chatcms.workspaceBackend = "gvisor"`
  in the NixOS module; switching is configuration only.
- **A measurement** of a large Hugo site's build time on gVisor against local
  processes, recorded in the design.

## Capabilities

### New Capabilities

- `workspace/gvisor`: workspaces in gVisor sandboxes.

### Modified Capabilities

## Impact

- `internal/workspace/gvisor`: the backend.
- `cmd/chatcms-relay`: the relay, built by Nix and found in the store.
- `internal/app`: choosing the backend; `CHATCMS_RUNSC`.
- `nix/nixos-module.nix`, `flake.nix`: the option, the packages, a check.
- No code outside `internal/workspace` changes for the workspace itself;
  the app only learns the backend's name.

## Assumptions

- **Rootless only.** Measured on 19 November 2026 with gVisor 20260406:
  rootless `runsc` has only the host's network, so isolation comes from
  running with none at all and passing sockets; `runsc do` cannot pass host
  sockets, an OCI bundle with `runsc run` can.
- **No dev mode on the host network**: the local backend serves development.
- **e2e on gVisor** cannot run inside the Nix build sandbox; the acceptance
  run on gVisor is a local run with `CHATCMS_WORKSPACE_BACKEND=gvisor`, like
  the micro VM's. Done on 19 November 2026 at 1440 px: 03 (forms and
  saving), 04 (pages and pictures), 05 (live preview), 33 (a preview of
  one's own), 35 (asking the assistant), 36 (a preview apart, PDF included),
  37 (select to edit), 46 (Chat + preview) and 48 (progress at the
  selection) pass; the PDF case of 36 once failed in a parallel run and
  passed alone, on both backends.
- **OpenCode** runs through glibc's loader in the sandbox
  (`opencode-sandboxed`), because gVisor's ELF loader refuses its binary.
- **`scripts/dev-be.sh`** gets `CHATCMS_DEV_BACKEND=gvisor`, to try it.
