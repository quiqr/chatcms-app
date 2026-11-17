# Tasks

## 1. The menu

- [x] 1.1 Remove the "Sign out everywhere" item from `PersonMenu.tsx`,
      `signOutEverywhere` from `web/src/lib/session.ts`, and the
      `person-menu.sign-out-everywhere` string from `en.ts` and `nl.ts`.
      Verify with `shells.test.tsx` asserting the menu holds "Sign out" and
      no "Sign out everywhere", and with `npm run typecheck` and
      `npm run check:strings`.

## 2. The server

- [x] 2.1 Remove `POST /api/sign-out/everywhere` from
      `internal/httpapi/router.go`, its handler and `SignOutEverywhere` from
      the `AuthService` interface in `internal/httpapi/auth.go`, and
      `Service.SignOutEverywhere` from `internal/auth/service.go`, keeping
      `Sessions.RevokeAll`. Replace the two endpoint tests in
      `auth_test.go` with one that the address now answers 404 or 405, and
      drop `TestSigningOutEverywhereEndsEveryoneOfYours`. Verify with
      `go test ./...` and `go test -tags=integration ./internal/auth/...`.

## 3. Documents and end to end

- [x] 3.1 Remove the "signing out everywhere" test from
      `test/e2e/14-signing-out.spec.mjs`. Verify that the e2e suite passes at
      1440 px and 390 px.
- [x] 3.2 Add a CHANGELOG entry under `[Unreleased]`, `### Removed`, at most
      two lines of at most 80 characters: "Sign out everywhere" is gone from
      the person menu. Verify by reading the section.
- [x] 3.3 Run `nix flake check` and verify it is green.

## Workflow follow-up

- Ship with `bash scripts/ship-change.sh sign-out-only "<subject>"`.
- Add the openspec-link to bean chatcms-ifc4 and close it with a Summary of
  Changes.
