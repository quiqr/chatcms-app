# Tasks

## 1. The version

- [x] 1.1 Add `Source`, `Changed()`, `Short()` and `Describe()` to
      `internal/version`, and make `String()` use `Describe()`. Verify with
      `internal/version` unit tests for a clean, a changed and an unknown
      source.
- [x] 1.2 Embed `LICENSE` in `version.go` beside `VERSION`. Verify with a
      unit test that the text starts with "Apache License".
- [x] 1.3 Let `mkServer` take an optional source, pass
      `self.shortRev or self.dirtyShortRev` in the `chatcmsd` package only,
      and point the e2e and NixOS VM checks at the `go-build` server. Verify
      with `nix build .#chatcmsd` and `result/bin/chatcmsd --version` naming
      the identifier, and the `go-build` binary saying "development build".

## 2. What it runs on

- [x] 2.1 Add `gitsync.GitVersion(ctx)`. Verify with a gitsync unit test
      against the pinned git that the answer starts with a digit.
- [x] 2.2 Keep the Hugo version found at start-up, add the PostgreSQL and Go
      questions, each with a two-second limit and "" on failure, behind an
      `About` source in `internal/app`. Verify with an integration test that
      the PostgreSQL version is not empty, and a unit test that a failing
      question leaves the others answered.

## 3. The answer

- [x] 3.1 Add `GET /api/administration/about` behind
      `ActionAdministerInstallation`, answering as in design.md. Verify with
      `internal/httpapi` tests: an administrator gets every field, somebody
      else is refused, and a method other than GET is refused.

## 4. The screen

- [x] 4.1 Add `web/src/routes/AboutScreen.tsx`, the route
      `/administration/about`, the About entry after History in
      `PlatformAdminShell.tsx`, and the strings in `en.ts` and `nl.ts`.
      Verify with Vitest: the version line for each case, "unknown" for an
      empty runtime version, the license text in its disclosure, the three
      source links and the contributors link opening outside, and the entry
      in the shell; and `npm run check:strings` and `npm run check:interface`.
- [x] 4.2 Add an e2e step to the platform admin spec: the administrator opens
      About, sees "development build" and the runtime versions, and opens
      the license; at 390 px nothing scrolls sideways. Verify with the e2e
      check at both widths.

## 5. Documents

- [x] 5.1 Add the About screen to `docs/design/DESIGN.md` beside the platform
      admin screens, and to `docs/DEPLOYING.md` as the place to see what is
      deployed. Verify that the prose hook passes on both.
- [x] 5.2 Add a CHANGELOG entry under `[Unreleased]`, `### Added`, at most two
      lines of at most 80 characters. Verify that the `changelog` check
      passes.
- [x] 5.3 Run `scripts/gate.sh release` once, because this change touches the
      flake and the VM check, and verify it is green.

## Workflow follow-up

- Ship with `bash scripts/ship-change.sh about-screen "<subject>"`.
- Add the openspec-link to bean chatcms-k6s0 and close it with a Summary of
  Changes.
- File a bean for the list of open source packages and their notices.
