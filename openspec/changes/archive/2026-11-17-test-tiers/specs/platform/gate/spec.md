## Purpose

Says which checks run at which moment, so that a change waits only for what
can break by it, and a release still waits for everything.

## ADDED Requirements

### Requirement: The checks come in three tiers

`scripts/gate.sh` SHALL take a tier and run its checks, stopping with a
failure when any fails. The `fast` tier SHALL run the build, the unit tests,
the frontend check, the sqlc check, the changelog check, the release check
and the dev scripts check. The `ship` tier SHALL run everything in `fast`, the
integration tests with the coverage gate, and the acceptance suite. The
`release` tier SHALL run `nix flake check`.

#### Scenario: A broken unit test

- **WHEN** a unit test fails and somebody runs `scripts/gate.sh fast`
- **THEN** the script fails and names the check that failed

#### Scenario: A broken acceptance test

- **WHEN** an acceptance test fails and somebody runs `scripts/gate.sh ship`
- **THEN** the script fails

#### Scenario: An unknown tier

- **WHEN** somebody runs `scripts/gate.sh quick`
- **THEN** the script fails and names the tiers there are

### Requirement: Shipping a change runs the ship tier

`scripts/ship-change.sh` SHALL run the `ship` tier and SHALL stop before
archiving, committing or pushing when it fails.

#### Scenario: The ship tier fails

- **WHEN** the acceptance suite fails while a change is shipped
- **THEN** nothing is archived, committed or pushed

#### Scenario: The deploy test is not waited for

- **WHEN** a change is shipped
- **THEN** the NixOS VM test is not run

### Requirement: A release runs every check

`scripts/release.sh` SHALL run the `release` tier, which is `nix flake check`
with every check, the NixOS VM test included.

#### Scenario: The deploy test fails at release

- **WHEN** the NixOS VM test fails during `scripts/release.sh minor`
- **THEN** nothing is committed, tagged or pushed
