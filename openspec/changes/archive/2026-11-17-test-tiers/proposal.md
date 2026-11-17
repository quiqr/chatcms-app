# Proposal

Bean: [chatcms-j31v](../../../.beans/chatcms-j31v--test-suite-tiers-and-a-parallel-e2e-suite.md)
(test suite tiers and a parallel e2e suite).

## Why

Shipping one change runs `nix flake check` in full: 165 acceptance tests one
after another, about 12 minutes, and then the NixOS VM deploy test. Most
changes touch neither the deploy module nor more than a few screens, yet each
waits for all of it. Every acceptance test already starts a ChatCMS,
PostgreSQL, remote and Mailpit of its own, so they can safely run side by
side; and the VM test checks something that changes far less often than a
feature.

## What Changes

- **The acceptance suite runs in parallel.** Several workers at once, with
  each test still owning everything it touches. The harness asks the system
  for free ports instead of picking random ones that can collide.
- **Three tiers**, run by one script, `scripts/gate.sh <tier>`:

  | Tier      | Checks                                                       |
  | --------- | ------------------------------------------------------------ |
  | `fast`    | go-build, go-test, frontend, sqlc, changelog, release, dev-scripts |
  | `ship`    | `fast` plus integration (with the coverage gate) and e2e     |
  | `release` | `nix flake check`: everything, the NixOS VM test included    |

- `scripts/ship-change.sh` runs the `ship` tier instead of `nix flake check`.
- `scripts/release.sh` keeps running `nix flake check`, now named the
  `release` tier.
- `fast` is for while working; nothing calls it on its own.
- `CLAUDE.md`, `docs/TESTING.md` and the dev shell greeting say which tier
  runs where.

## Capabilities

### New Capabilities

- `platform/gate`: which checks run in which tier, and which tier guards
  shipping a change and making a release.

### Modified Capabilities

- `acceptance/harness`: the suite runs its tests side by side, and a test's
  ports are free ones the system hands out.

## Impact

- `test/e2e/playwright.config.mjs` and `test/e2e/harness.mjs`.
- New `scripts/gate.sh` with `scripts/gate.test.sh`, shellchecked and run by
  the `dev-scripts` check; `scripts/ship-change.sh` calls it.
- `CLAUDE.md` (the gate section), `docs/TESTING.md` ("What the gate runs"),
  the dev shell greeting in `flake.nix`.
- No application code changes. No seam is touched.
- Acceptance tests: all nine keep running on every shipped change; only the
  NixOS VM test, behind the "builds and deploys via the NixOS module" line of
  section 6, moves to the release tier.

## Assumptions

- A change can reach `main` that breaks the NixOS VM test, and is caught at
  the next release, before any version is tagged. That trade is what Pim
  chose: the module changes rarely, and a release cannot go out without it.
- The worker count defaults to half the available cores, at least one, and
  can be set with `E2E_WORKERS`. Inside the Nix build that is what the
  sandbox reports.
- Retries stay at zero: a test that fails only when run beside others is a
  bug to fix, not to retry.
- `nix flake check` stays the full gate and keeps every check, so running it
  by hand still means what it meant.
