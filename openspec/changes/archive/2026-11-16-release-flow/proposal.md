# Proposal

Bean: [chatcms-y0yk](../../../.beans/chatcms-y0yk--release-flow.md) (release
flow), milestone chatcms-27ng (05 develop happiness). Adapted from the nivis
`release-management` change.

## Why

ChatCMS has no releases. The version is written into `flake.nix` by hand as
`0.1.0` while nothing has been released, a plain `go build` reports `dev`,
and `CHANGELOG.md` only ever grows an `[Unreleased]` section. A machine that
deploys ChatCMS through the flake has nothing better to pin than a commit.

## What Changes

- A top-level `VERSION` file, starting at `0.0.0`, is the only place the
  version is written. `flake.nix` reads it, and the server reports it from
  every build path.
- `scripts/release.sh patch|minor|major [--dry-run]` cuts a release. It:
  - refuses on a dirty working copy, or when `main` differs from the remote
  - bumps `VERSION`
  - moves `[Unreleased]` in `CHANGELOG.md` into a dated section
  - runs `nix flake check`
  - commits, tags `v<version>`, and pushes `main` and the tag
  - creates a GitHub release with that changelog section as its notes
- `docs/RELEASING.md` describes the flow and how a machine pins a release.
  The README links to it.
- The first release (`v0.1.0`) is not cut by this change. It is a separate
  step, triggered by hand after the change ships.

## Capabilities

### New Capabilities

- `platform/release`: how the version is recorded, how a release is cut, and
  what a release is.

### Modified Capabilities

None.

## Impact

- New: `VERSION`, `scripts/release.sh`, its test, `docs/RELEASING.md`.
- Changed: `flake.nix` (version from `VERSION`, a check for the script),
  `internal/version` (default from the embedded `VERSION`), `README.md`,
  `CHANGELOG.md`.
- No seam (gitsync, runner, policy, auth) is touched.

## Acceptance test served

None from `docs/BRIEFING-M01.md` section 6 directly. It gives the first
customer deployment a version to pin.

## Assumptions

- A release is the tagged flake: `github:quiqr/chatcms-app/v0.1.0`. No
  binaries are attached, because the server only carries its frontend when
  Nix builds it, and deployment is Nix-only.
- No CI. The GitHub release is made by `gh` from the machine that cuts it,
  so `gh` must be signed in there.
- `jj` makes the release commit, like `scripts/ship-change.sh`. The tag is a
  git tag in the colocated repository.
- The version follows semver. Before 1.0, `minor` is used for features and
  `patch` for fixes.
