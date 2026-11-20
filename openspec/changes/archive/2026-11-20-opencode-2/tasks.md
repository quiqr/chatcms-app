# Tasks

## 1. The client

- [x] 1.1 Session, prompt, interrupt and events on OpenCode 2's routes;
  the relay translates the v2 events; proven by Go tests against a fake v2
  server.
- [x] 1.2 A refusal or a third attempt ends the answer with the reason and
  interrupts the session; proven by a Go test.

## 2. Packaging

- [x] 2.1 `flake.nix` takes OpenCode from the main nixpkgs and drops
  `nixpkgs-opencode`; the lock follows.
- [x] 2.2 e2e `35-asking-the-assistant` and the other assistant specs pass
  with OpenCode 2 (16 passed); the gVisor and deploy VM tests and the
  workspace guest build pass. The hand-run micro VM boot test was not run:
  it needs a tap device and root on this machine.

## 3. Ship

- [x] 3.1 `CHANGELOG.md` under [Unreleased].
- [x] 3.2 Ship; close chatcms-uzi2.
