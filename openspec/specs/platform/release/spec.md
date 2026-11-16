# platform/release Specification

## Purpose
Defines where the version of ChatCMS is recorded, how a release is cut, and
what a release consists of, so that a deployment can pin a known version.

## Requirements

### Requirement: The version is written in one place

The version SHALL be written only in the top-level `VERSION` file, as a bare
semver. The package version in the flake and the version the server reports
SHALL both come from it.

#### Scenario: A Nix build reports the version

- **WHEN** `VERSION` holds `0.3.1` and the server is built with Nix
- **THEN** the package version SHALL be `0.3.1`
- **AND** the server SHALL report `0.3.1`

#### Scenario: A plain Go build reports the version

- **WHEN** `VERSION` holds `0.3.1` and the server is built with `go build`
  and no linker flags
- **THEN** the server SHALL report `0.3.1`

### Requirement: A release bumps the version and dates the changelog

`scripts/release.sh` SHALL take `patch`, `minor` or `major`, write the next
semver to `VERSION`, and turn the `[Unreleased]` section of `CHANGELOG.md`
into a section for that version dated today, above a new empty
`[Unreleased]`.

#### Scenario: A minor release

- **WHEN** `VERSION` holds `0.1.4` and `scripts/release.sh minor` runs
- **THEN** `VERSION` SHALL hold `0.2.0`
- **AND** `CHANGELOG.md` SHALL hold an empty `[Unreleased]` section followed
  by `[0.2.0]` with today's date and what `[Unreleased]` held

#### Scenario: A major release resets minor and patch

- **WHEN** `VERSION` holds `0.4.2` and `scripts/release.sh major` runs
- **THEN** `VERSION` SHALL hold `1.0.0`

#### Scenario: A dry run changes nothing

- **WHEN** `scripts/release.sh minor --dry-run` runs
- **THEN** it SHALL print the next version and the changelog section it would
  release
- **AND** no file SHALL change and nothing SHALL be committed, tagged or
  pushed

### Requirement: A release refuses an unsafe starting point

`scripts/release.sh` SHALL refuse, before changing anything, when the working
copy has changes, when `main` differs from the remote, when `[Unreleased]` is
empty, when the tag for the next version already exists, or when `gh` is not
signed in.

#### Scenario: Uncommitted changes

- **WHEN** the working copy holds a change and `scripts/release.sh patch` runs
- **THEN** it SHALL exit with a failure naming the reason
- **AND** `VERSION` and `CHANGELOG.md` SHALL be unchanged

#### Scenario: Nothing to release

- **WHEN** `[Unreleased]` holds no entries
- **THEN** it SHALL exit with a failure naming the reason

### Requirement: A release is gated, tagged and published

After writing `VERSION` and `CHANGELOG.md`, `scripts/release.sh` SHALL run
`nix flake check` and stop on failure without committing. On success it SHALL
commit the two files, tag the commit `v<version>`, push `main` and the tag,
and create a GitHub release for the tag whose notes are the released changelog
section.

#### Scenario: The gate fails

- **WHEN** `nix flake check` fails during a release
- **THEN** nothing SHALL be committed, tagged or pushed

#### Scenario: A release is published

- **WHEN** the gate passes during `scripts/release.sh minor` from `0.1.4`
- **THEN** a commit SHALL change `VERSION` and `CHANGELOG.md` only
- **AND** tag `v0.2.0` SHALL point at that commit on the remote
- **AND** a GitHub release `v0.2.0` SHALL carry the `[0.2.0]` section as notes
