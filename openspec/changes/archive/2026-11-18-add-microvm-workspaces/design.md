# Design

## Context

See proposal.md. `internal/workspace` has a `Backend` seam and one backend,
`local`, which runs Hugo as a child process on 127.0.0.1. Each person has a
copy of the site on the host (`<sites>/<project>/people/<person>/working-copy`).
Builds run Hugo through `runner.Sandbox` (bubblewrap, `/nix/store` bound
read-only). A project records `hugo_version` (templates, project settings).

Measured on the development machine (i7-1250U), the guest from
`nix/workspace-guest.nix` boots on Cloud Hypervisor 52 and answers on port
1313 after 4.8 to 6.2 s cold through the backend (5.3 s by hand): about 3.6 s to the guest's second-stage systemd
(kernel and initrd), the rest to Hugo. Writing the kernel console to the
emulated serial port cost seconds, so the guest logs to the virtio console.

## Goals / Non-Goals

**Goals:** a `microvm` backend behind the existing seam, with nothing outside
`internal/workspace` aware of it; a project's own Hugo in previews and builds.

**Non-Goals:** boot under 3 s (a trimmed kernel or snapshot restore, filed as
its own bean); VMs surviving a restart; writes from inside the VM; vsock.

## Decisions

### 1. One guest, started by the backend, not by microvm.nix's runner

microvm.nix builds the guest (`nix/workspace-guest.nix`): kernel, initrd and
command line, written to `chatcms-workspace-guest.json` by the flake package
`workspace-guest`. Its runner script fixes shares and interfaces at build
time, so the backend starts each VM itself:

- two `virtiofsd --readonly --sandbox=none` processes: tag `ro-store` on
  `/nix/store` (`--cache=auto`) and tag `site` on the person's copy
  (`--cache=never`, so a save shows at once);
- `cloud-hypervisor --kernel --initramfs --cmdline … --memory
  size=1024M,shared=on --fs … --net tap=<tap>,mac=<mac> --serial off
  --console file=<dir>/console.log`.

Per-workspace values travel on the kernel command line: `chatcms.ip`,
`chatcms.gateway`, `chatcms.hugo`, `chatcms.under`, `chatcms.reload`. The
guest's `chatcms-preview` unit starts straight after the file systems, with
no default dependencies, configures the interface and runs Hugo as uid 65534
with its cache and resources under `/run/preview`, because `/site` is
read-only. Hugo watches by polling (`--poll 500ms`): a write on the host sends
no inotify event into the guest. A save shows in about 0.6 s.

Sockets live in a short directory (`CHATCMS_WORKSPACE_RUN_DIR`, default
`/run/chatcms/w`): a socket path must stay under 108 bytes.

*Alternative:* microvm.nix's host module, one systemd unit per VM. Declared
at build time; workspaces come and go at run time.

### 2. A pool of TAP devices, made by the NixOS module

Creating a TAP needs `CAP_NET_ADMIN`; using one owned by the service user does
not. The module creates `chv0 … chv<N-1>` (N = `workspacesAtOnce`) owned by
the service user, host address `10.231.<i>.1/30`, guest `10.231.<i>.2`, MAC
`02:00:0a:e7:<i>:02`. nftables drops forwarding from `chv*` and accepts
nothing on input from them but replies. The backend hands out free slots under
a mutex; none free is `workspace.ErrTooMany`. `scripts/dev-taps.sh up|down`
does the same with sudo for development.

### 3. Handles and stopping

The handle is `<slot>:<ch pid>:<store pid>:<site pid>`. `Stop` sends SIGTERM to
cloud-hypervisor and both virtiofsd processes, SIGKILL after 5 s, removes the
socket directory and frees the slot; asking twice does nothing. `Apart()` is
false. `Alive` checks the cloud-hypervisor pid's command line names the
configured binary and that `10.231.<i>.2:1313` answers; an adopted workspace
takes its slot back.

### 4. A project's own Hugo: `internal/nixpin`

- `Resolve(ctx, name, version)`: `GET <nixhub>/v2/resolve?name=hugo&version=…`
  gives a nixpkgs revision and attribute path. Base URL configurable
  (`CHATCMS_NIXHUB_URL`, default `https://search.devbox.sh`), for tests.
- `Realise`: `nix build --print-out-paths --out-link <state>/gcroots/hugo-<version>
  github:NixOS/nixpkgs/<rev>#<attr>`, so the path is a garbage collector root.
- Remembered in `nix_packages (name, version, revision, attr, store_path,
  realised_at)`; a later ask with the store path still present asks nothing
  outside. Concurrent asks for one version wait for the first.
- Unknown version: `refused.hugo-version-not-found` with the version.
- The app's `hugoFor(project)` is the installation Hugo when the project
  records no version, else `nixpin` for `hugo` at that version, binary
  `<path>/bin/hugo`.

`workspace.Spec` gains `Hugo`; the manager fills it through
`Settings.HugoFor`; the local backend uses it when set. `sitebuild` takes the
Hugo per build (`Options.Hugo`), and the sandbox already binds `/nix/store`.

### 5. Configuration

`CHATCMS_WORKSPACE_BACKEND` (`local`|`microvm`), `CHATCMS_CLOUD_HYPERVISOR_BIN`,
`CHATCMS_VIRTIOFSD_BIN`, `CHATCMS_WORKSPACE_GUEST` (the JSON),
`CHATCMS_WORKSPACE_TAPS` (count), `CHATCMS_NIX_BIN`, `CHATCMS_NIXHUB_URL`. The
NixOS module's `workspaceBackend = "microvm"` sets them, creates the taps and
firewall, lets the service reach `/dev/kvm` and `/dev/net/tun`
(`PrivateDevices = false`, `DeviceAllow`, group `kvm`) and adds a
`RuntimeDirectory`.

## Risks / Trade-offs

- [Nixhub is a service outside the installation] → results are remembered;
  only a version nobody asked for before needs it.
- [5 s cold start] → the workspace starts when the project is opened; boot
  work is its own bean.
- [The gate has no KVM] → backend logic is unit-tested with fake processes;
  a `microvm`-tagged integration test boots a real VM and is run by hand.
- [Hugo writing into a read-only site] → cache and resources go to `/run`.

## Migration Plan

Nothing changes until `workspaceBackend = "microvm"` is set. Rollback is
setting it back.
