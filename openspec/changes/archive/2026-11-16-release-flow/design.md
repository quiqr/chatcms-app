# Design

## Context

`flake.nix` sets `version = "0.1.0"` and passes it to the server with
`-X internal/version.Version`. `internal/version.Version` defaults to `dev`.
`scripts/ship-change.sh` commits with `jj commit`, moves `main` to `@-` and
pushes it. There is no CI.

## Goals / Non-Goals

**Goals:**

- Cutting a release is one command, and `--dry-run` shows what it would do.
- The release commit holds only `VERSION` and `CHANGELOG.md`.

**Non-Goals:**

- goreleaser, attached binaries, CI, signing.
- Cutting `v0.1.0`; that is done by hand after this ships.

## Decisions

### `VERSION` is read by the flake and embedded into Go

`flake.nix` uses `lib.fileContents ./VERSION`. A small root-level package
`github.com/quiqr/chatcms-app` with `//go:embed VERSION` exposes it, because
`go:embed` cannot reach a parent directory from `internal/version`.
`internal/version` defaults `Version` to it. The ldflags stay, so a Nix build
and a plain build agree by construction and the flag can override for a
one-off build.

Alternative: generate a Go file from `VERSION`. That adds a step which can
drift, which is the problem being solved.

### The release script follows `ship-change.sh`

The order is: checks, write both files, `nix flake check`, `jj commit`,
`jj bookmark set main -r @-`, `git tag v<x> <commit>`, `jj git push --bookmark main`,
`git push origin v<x>`, `gh release create`. The tag is a git tag, because
the repository is colocated and a git tag is what GitHub and flake
references read.

If the gate fails, the script restores both files, so a failed release leaves
the working copy as it found it.

### The changelog section is moved with awk

The section between `## [Unreleased]` and the next `## [` is extracted, both
for the notes and for the rewrite. The link references at the bottom of a
Keep a Changelog file are not maintained. This changelog has none.

### The script is tested with stubs, like the dev scripts

`scripts/release.test.sh` copies the script into a temporary git repository
with a `VERSION` and a `CHANGELOG.md`, and puts stub `nix`, `jj`, `gh` and
`git push` on `PATH`. It checks the bump, the changelog rewrite, the dry run,
each refusal, and that a failing stub `nix` leaves the files untouched. It
runs in a `release` flake check together with `shellcheck`.

## Risks / Trade-offs

- [`gh` is not signed in, so the release is tagged but not published] →
  The script checks `gh auth status` before it changes anything.
- [The push succeeds but `gh release create` fails] → The script prints the
  exact `gh release create` command to rerun. The tag is the release that
  matters for flakes.
