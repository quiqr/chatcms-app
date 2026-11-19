# Design

## Context

`workspace.Backend` has `Start`, `Alive`, `Stop`. The micro VM backend boots
a guest whose scripts run the preview command (passed as base64 JSON `run`:
command, PATH, env) and OpenCode with the AI gateway's address and token;
the host reaches them on the guest's address. The local backend runs the
same as host processes.

## The spike (19 November 2026)

- `runsc --rootless` works on this host; `--network=sandbox` falls back to
  the host network rootless, which reaches the internet.
- `--network=none` with `--host-uds=all` passes host Unix sockets both ways,
  but only with `runsc run` on an OCI bundle: `runsc do` always overlays the
  root and cannot.
- Socket paths must stay under 108 bytes, so sockets live in a short
  directory per workspace (`/run/user/<uid>/chatcms/<id>` or the state
  directory's `s/<short id>`).

## Decisions

### 1. One bundle per workspace

`config.json`: root an empty read-only directory; mounts `/nix/store` (ro),
`/site` (rw, the person's copy), `/tmp` (tmpfs), `/run/chatcms` (rw, the
socket directory), `/proc`; namespaces pid, ipc, uts, mount; the process a
launcher from the store (`chatcms-relay launch`) with the run JSON and the
token in its environment. `runsc --rootless --network=none --host-uds=all
run --bundle <dir> <id>`, detached; the handle holds the sandbox id and the
host relay ports.

### 2. The relay

`chatcms-relay` is one small Go program with three verbs:
- `out <unix> <tcp>`: inside, listen on the Unix socket, forward to a TCP
  port inside (preview, agent).
- `in <tcp> <unix>`: inside, listen on a TCP port, forward to a Unix socket
  of the host (the AI gateway).
- `launch`: inside, start the preview command and OpenCode as the micro VM
  guest does, the relays beside them; the sandbox ends when the preview
  ends, and an agent that ends is logged while the preview stays.
The host side is goroutines in the backend: a local TCP listener per
workspace for preview and agent forwarding to the sockets, and a Unix
listener for the gateway forwarding to its TCP address.

### 3. Lifecycle

`Alive` asks `runsc state <id>` and checks the host relays; `Stop` runs
`runsc kill` and `runsc delete --force`, closes the host relays and removes
the bundle and sockets. On server start, leftover sandboxes under the state
root are deleted, as the micro VM backend reconciles.

### 4. Configuration

`CHATCMS_WORKSPACE_BACKEND=gvisor`, `CHATCMS_RUNSC` (the runsc binary),
`CHATCMS_RELAY` (the relay in the store). The NixOS module sets them from
`workspaceBackend = "gvisor"` and puts gvisor and the relay in the closure.

### 5. OpenCode through its loader

gVisor's own ELF loader refuses OpenCode's Bun-compiled binary ("exec format
error"), while the binary runs when glibc's dynamic loader loads it. The
flake builds `opencode-sandboxed`: a script that puts ripgrep on the PATH, as
the package's wrapper does, and runs the binary through the loader. The
module gives it to the gVisor backend.

## Measured (19 November 2026)

A generated Hugo site of 3000 pages (12 MB of Markdown), Hugo 0.168, a full
build, three runs each on the same host:

| Where                 | Seconds   |
| --------------------- | --------- |
| Local process         | 0.41-0.43 |
| gVisor sandbox        | 1.91-2.28 |
| Empty gVisor sandbox  | 0.10-0.13 |

A full build is about 4 to 5 times slower under gVisor, from file access.
A preview pays it on its first build; Hugo's server rebuilds only what
changed after that. A workspace starts in about half a second.

## Risks / Trade-offs

- **File I/O** is slower under gVisor; measured on a large Hugo site and
  written here before shipping.
- **Isolation** is weaker than a VM (a user-space kernel, not hardware) and
  stronger than processes; with no network at all it is stricter than the
  briefing's "ordinary container networking".
- **Host sockets** depend on `--host-uds`, a runsc feature that could change;
  the version is pinned through nixpkgs.
