# Design

## Context

`flake.nix` defines the checks `go-build`, `go-test`, `integration`,
`frontend`, `sqlc`, `dev-scripts`, `release`, `changelog`, and on Linux
`nixos-vm` and `e2e`. `scripts/ship-change.sh` and `scripts/release.sh` both
run `nix flake check`. The e2e check runs Playwright with `workers: 1` and
`fullyParallel: false`; each test's fixture starts its own `Harness`, which
puts PostgreSQL on a socket in its own directory but picks the server port as
`18080 + random(2000)` and derives Mailpit's ports from it.

## Goals / Non-Goals

**Goals:**

- The `ship` tier finishes in a few minutes on a developer's machine.
- One place names which check belongs to which tier.

**Non-Goals:**

- Choosing e2e specs by what a change touched. It is fragile, and parallel
  runs make it unnecessary.
- Changing what any check tests.

## Decisions

### Parallel within the suite

`playwright.config.mjs` sets `fullyParallel: true` and
`workers: Number(process.env.E2E_WORKERS) || Math.max(1, Math.floor(os.availableParallelism() / 2))`.
Half the cores, because every test runs a browser, a Go server, PostgreSQL,
Mailpit and Hugo builds; more than that makes timeouts the bottleneck.
`fullyParallel` spreads the tests of one large file over workers too, which
keeps the run from waiting on the longest file.

Alternative: shard into several Nix derivations. Rejected: each shard would
rebuild its node_modules and start its own browser install, and the output of
one failing shard is harder to find.

### Free ports

`Harness` gets a `freePort()` that listens on port 0 on 127.0.0.1, reads the
port it was given and closes it. The server, the SMTP and the inbox port each
get one. A port can in theory be taken between closing and the server
binding it; the harness already waits for the server to answer and fails with
its log, so such a collision shows as a clear start failure rather than a
wrong answer, and with ports handed out by the system it is rare enough not
to retry.

### One gate script

`scripts/gate.sh <fast|ship|release>`:

- reads the system once with `nix eval --impure --raw --expr builtins.currentSystem`;
- for `fast` and `ship`, runs one `nix build --no-link -L` over
  `.#checks.<system>.<name>` for every check of the tier, so Nix builds them
  side by side and reuses what is cached;
- for `release`, runs `nix flake check`;
- for anything else, prints the three tiers and exits with 2.

The program it calls is `${NIX:-nix}`, so `scripts/gate.test.sh` can put a
stub in its place and check which checks each tier asks for, the same way the
release script is tested. The test runs inside the `dev-scripts` check, which
also shellchecks the new script.

`ship-change.sh` calls `scripts/gate.sh ship` in place of `nix flake check`.
`release.sh` keeps calling `nix flake check`: that is the `release` tier, and
leaving it alone keeps `release.test.sh` as it is.

### The rule in the documents

`CLAUDE.md` says the gate for a change is `scripts/gate.sh ship`, the gate
for a release is `nix flake check`, and neither may be bypassed. The
`docs/TESTING.md` table gains a tier column.

## Risks / Trade-offs

- [A change that breaks the deploy module reaches `main`] → `release.sh`
  runs the VM test before anything is tagged; anybody touching
  `nix/nixos-module.nix` can run `scripts/gate.sh release` by hand, and
  `CLAUDE.md` says so.
- [Parallel runs expose tests that leak into each other] → every test owns
  its world already; a test that fails only in company is a real bug and is
  fixed, not retried.
- [A loaded machine makes a timeout fire] → timeouts stay as they are, and
  `E2E_WORKERS` lowers the count where a machine is small.

## Migration Plan

None. The next ship uses the new tier; a release behaves as before.
