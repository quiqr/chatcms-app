# Proposal

Bean: [chatcms-7p90](../../../.beans/chatcms-7p90--micro-vm-workspace-backend.md)
(micro VM workspace backend), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).
It absorbs chatcms-90m8 (build with a project's own Hugo version).

## Why

A workspace is about to run an agent that executes code nobody reviewed. The
local backend runs Hugo as a process of the server, with the server's view of
the machine. Each workspace needs a machine of its own, and each site needs
the Hugo it was written for rather than the one the installation happens to
pin.

## What Changes

- **A `microvm` workspace backend.** Each workspace is a Cloud Hypervisor
  micro VM booted from one NixOS guest built with microvm.nix. The host
  `/nix/store` is shared read-only over virtio-fs; the person's copy of the
  site is shared over virtio-fs as a second, read-only mount.
- **A TAP device and a private address per VM**, from a pool the NixOS module
  creates. The VM reaches nothing but the host: no forwarding, and the host
  answers it on no port yet.
- **The preview proxy goes to the VM's address and port**, readiness by TCP,
  live reload as before.
- **A project's own Hugo version.** The version a project records (set from
  its template or in its settings) is resolved to a nixpkgs revision through
  the Nixhub API, realised with `nix build`, and remembered in Postgres.
  Previews and one-shot builds both run that Hugo. A project without one uses
  the installation's Hugo.
- **`CHATCMS_WORKSPACE_BACKEND`** chooses `local` (default) or `microvm`; the
  NixOS module gains `workspaceBackend` and the tap pool.

## Capabilities

### New Capabilities

- `workspace/microvm`: what a micro VM workspace is, what it can see and
  reach, and how it is started and stopped.
- `workspace/hugo-version`: how a project's own Hugo version is found, made
  available and used, and what happens when it cannot be.

### Modified Capabilities

- `preview/serving`: a preview is built by the project's own Hugo when it
  names one, rather than always by the installation's.

## Impact

- New `internal/workspace/microvm`, `internal/nixpin`; `internal/workspace`
  gains the Hugo in `Spec`; `internal/sitebuild` takes the Hugo per build.
- Seams: `workspace.Backend` (the point of the epic), `runner.Runner`
  untouched (builds pass a different Hugo path through it). `gitsync` gains
  nothing.
- `nix/`: the guest (`nix/workspace-guest.nix`), the tap pool and firewall in
  the NixOS module, a flake input for microvm.nix, a check that builds the
  guest. Migration `00019_nix_packages.sql`.
- Acceptance (milestone 07, epic 3): a preview runs inside a VM with the
  project's pinned Hugo version.

## Assumptions

- **The person's copy is shared into the VM, not cloned and pushed back.**
  The briefing has the VM clone a bare repo and push to it. Since
  chatcms-gkdd every person already has a copy on the host that forms write
  to; sharing its tree read-only gives the VM the site with no git, no clone
  on a cold start, and no second copy to keep in step. The chat epic, which
  writes from inside the VM, decides between a writable share of the tree
  (with the git directory kept outside it) and clone and push.
- **The pinned version lives in the project record**, where templates and
  project settings already put it, not in the site's quiqr config. The
  runtime definition epic moves it into the site.
- **Micro VMs stop with the app**, like the local backend: cloud-hypervisor
  and virtiofsd run as children of the server. Running them as units of their
  own, so they survive a restart, is left for when restarts hurt.
- **The gate cannot boot a VM**: the Nix build sandbox has no KVM or TAP
  devices. The gate builds the guest and unit-tests the backend; booting a
  real VM is an integration test that runs where KVM and the tap pool exist,
  by hand before shipping, and its result is written in the archive summary.
