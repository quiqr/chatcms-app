# Tasks

## 1. The suite in parallel

- [x] 1.1 Give `test/e2e/harness.mjs` a `freePort()` and use it for the
      server, SMTP and inbox ports instead of the random range. Verify by
      running the e2e check and seeing every spec pass.
- [x] 1.2 Set `fullyParallel: true` and the worker count from `E2E_WORKERS`
      or half the cores in `test/e2e/playwright.config.mjs`. Verify by running
      `nix build .#checks.x86_64-linux.e2e -L` and seeing all 165 tests pass,
      noting the time against the 12 minutes before; and once with
      `E2E_WORKERS=1` locally on one spec to see the one-at-a-time path.

## 2. The tiers

- [x] 2.1 Add `scripts/gate.sh` with the three tiers as in design.md, calling
      `${NIX:-nix}`. Verify with `scripts/gate.test.sh`, using a stub for
      `nix`: `fast` asks for exactly its seven checks, `ship` for those plus
      `integration` and `e2e`, `release` runs `flake check`, a failing check
      fails the script, and an unknown tier exits 2 and names the tiers.
- [x] 2.2 Add `gate.sh` and `gate.test.sh` to the `dev-scripts` check in
      `flake.nix`, shellchecked and run. Verify with
      `nix build .#checks.x86_64-linux.dev-scripts`.
- [x] 2.3 Make `scripts/ship-change.sh` run `scripts/gate.sh ship` in place of
      `nix flake check`. Verify by shipping this change, which runs it.

## 3. Documents

- [x] 3.1 Update the gate section of `CLAUDE.md`, "What the gate runs" in
      `docs/TESTING.md` with a tier column, and the dev shell greeting in
      `flake.nix`. Verify that the prose hook passes on both documents.
- [x] 3.2 Add a CHANGELOG entry under `[Unreleased]`, `### Changed`, at most
      two lines of at most 80 characters: shipping a change waits for the ship
      tier, with the acceptance tests in parallel; a release still runs every
      check. Verify that the `changelog` check passes.

## 4. The whole

- [x] 4.1 Run `scripts/gate.sh release` once, which is `nix flake check`, and
      verify it is green.

## Workflow follow-up

- Ship with `bash scripts/ship-change.sh test-tiers "<subject>"`.
- Add the openspec-link to bean chatcms-j31v and close it with a Summary of
  Changes.
