# Tasks

## 1. The guest

- [x] 1.1 `nix/workspace-guest.nix` and the flake package and check
  `workspace-guest`; proven by `nix build .#workspace-guest` and the check in
  the ship tier.
- [x] 1.2 Boot measured by hand (cold and warm) and written in design.md;
  a bean for faster boots filed.

## 2. A project's own Hugo

- [x] 2.1 Migration `00019_nix_packages.sql` and queries; `sqlc diff` clean.
- [x] 2.2 `internal/nixpin`: resolve through Nixhub, realise with `nix
  build`, remember, one ask per version at a time; proven by unit tests with
  an `httptest` Nixhub and a fake `nix`, and an integration test of the
  Postgres store.
- [x] 2.3 An unknown version is refused with `refused.hugo-version-not-found`
  naming it; strings in `en.ts` and `nl.ts`; proven by a unit test.
- [x] 2.4 `workspace.Spec.Hugo` through `Settings.HugoFor`; the local backend
  runs it; proven by manager and local tests.
- [x] 2.5 `sitebuild` takes the Hugo per build and the publish flow passes the
  project's; proven by sitebuild tests.

## 3. The microvm backend

- [x] 3.1 `internal/workspace/microvm`: slot pool, command lines, start,
  readiness, stop, alive; proven by unit tests with fake binaries.
- [x] 3.2 `internal/workspace/microvm` integration test (`microvm` tag) that
  boots the real guest and serves a site; run by hand, result in the archive
  summary.
- [x] 3.3 `CHATCMS_WORKSPACE_BACKEND` and the other variables; proven in
  `config_test.go`.

## 4. Deployment

- [x] 4.1 The NixOS module: `workspaceBackend`, taps, nftables, device access,
  runtime directory; `scripts/gate.sh release` green.
- [x] 4.2 `scripts/dev-taps.sh up|down`, shellchecked by the `dev-scripts`
  check.
- [x] 4.3 `docs/ARCHITECTURE.md` and `docs/DEPLOYING.md`.

## 5. Ship

- [x] 5.1 `CHANGELOG.md` under [Unreleased].
- [x] 5.2 `scripts/gate.sh ship` green.
