---
# chatcms-j31v
title: test suite tiers and a parallel e2e suite
status: completed
type: task
created_at: 2026-11-17T21:43:44Z
updated_at: 2026-11-18T01:45:00Z
openspec-link: openspec/changes/archive/2026-11-17-test-tiers
---

The gate runs the whole e2e suite (165 tests, one at a time, about 12 minutes) and the NixOS VM test on every shipped change. That is too long to wait for at every change.

- Run the e2e tests in parallel; each already has a ChatCMS of its own.
- Tiers: a fast tier while working, a ship tier for every change, and the full gate for a release.

## Summary of Changes

Shipped with `test-tiers` (commit ae40868):

- The acceptance suite runs its tests side by side, half the cores by
  default or `E2E_WORKERS`; the harness takes free ports from the system.
  165 tests at both widths went from about 12 minutes to about 3.
- `scripts/gate.sh fast|ship|release` runs a tier of the checks, with its own
  test in the `dev-scripts` check.
- `scripts/ship-change.sh` runs the ship tier, which leaves out only the NixOS
  VM test; `scripts/release.sh` still runs `nix flake check`.
- `CLAUDE.md`, `docs/TESTING.md` and the dev shell greeting name the tiers.
