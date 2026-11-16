# Tasks

## 1. One version

- [x] 1.1 Add `VERSION` holding `0.0.0` and a root package that embeds it. Verify with a Go test that the embedded value matches the file, trimmed.
- [x] 1.2 Default `internal/version.Version` to the embedded `VERSION`. Verify with the existing version test, extended: without ldflags, `String()` carries `0.0.0`.
- [x] 1.3 Read `VERSION` in `flake.nix` in place of the hardcoded `0.1.0`. Verify that `nix build .#chatcmsd` and `result/bin/chatcmsd --version` report `0.0.0`.

## 2. The release script

- [x] 2.1 Add `scripts/release.sh patch|minor|major [--dry-run]` with the bump and the changelog rewrite. Verify with `scripts/release.test.sh`: patch, minor and major from `0.1.4`, the dated section, and a new empty `[Unreleased]`.
- [x] 2.2 Add the refusals: dirty working copy, `main` not equal to the remote, empty `[Unreleased]`, existing tag, `gh` not signed in. Verify with `scripts/release.test.sh`: each refuses and leaves the files unchanged.
- [x] 2.3 Add the gate, commit, tag, push and `gh release create` with the section as notes. Verify with `scripts/release.test.sh`: stubs record the calls in order, a failing `nix` stops before `jj` and restores the files, and `--dry-run` calls none of them.
- [x] 2.4 Add a `release` check to `flake.nix` that runs `shellcheck` on the script and runs `scripts/release.test.sh`. Verify with `nix build .#checks.x86_64-linux.release`.

## 3. Docs

- [x] 3.1 Write `docs/RELEASING.md`: when to use patch, minor or major, what the script does, what to do when `gh` fails, and how a machine pins `github:quiqr/chatcms-app/v<x>`. Link it from the README and add `release.sh` to the command list in `AGENTS.md` and `CLAUDE.md`. Verify by running the documented `--dry-run` command.
- [x] 3.2 Add the release flow to `CHANGELOG.md` under `[Unreleased]`. Verify by reading the entry.

## Workflow follow-up

- Ship with `scripts/ship-change.sh release-flow "<subject>"`.
- Add `openspec-link` to bean chatcms-y0yk.
- Cut the first release by hand: `scripts/release.sh minor --dry-run`, then `scripts/release.sh minor` for `v0.1.0`. Close the bean after that.
