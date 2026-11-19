# Tasks

## 1. The relay

- [x] 1.1 `cmd/chatcms-relay` with `out`, `in` and `launch`; proven by Go tests
  (forwarding both ways, launch stops all when one stops).
- [x] 1.2 Built by Nix as `chatcms-relay`.

## 2. The backend

- [x] 2.1 `internal/workspace/gvisor`: bundle, start, host relays, `Alive`,
  `Stop`, leftovers; proven by Go tests with a fake runsc.
- [x] 2.2 A boot test with the real runsc (skipped where it is absent): the
  preview answers through its socket; the sandbox reaches nothing outside;
  stopping leaves nothing.
- [x] 2.3 Chosen by `CHATCMS_WORKSPACE_BACKEND=gvisor` in the app.

## 3. The installation

- [x] 3.1 NixOS module option and environment; the flake builds and checks.
- [x] 3.2 Measure a large Hugo site's build on gVisor and locally; record it.

## 4. Proof and ship

- [x] 4.1 The acceptance run with `CHATCMS_WORKSPACE_BACKEND=gvisor`, recorded.
- [x] 4.2 `CHANGELOG.md` under [Unreleased].
- [x] 4.3 `scripts/gate.sh release` (nix/ changes) green; close chatcms-7l81.
