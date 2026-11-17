# Proposal

Bean: [chatcms-k6s0](../../../.beans/chatcms-k6s0--about-screen-in-administration.md)
(about screen in administration).

## Why

Whoever runs an installation cannot see from the interface what it is
running. The version exists only behind `chatcmsd --version`, and it says
"0.1.0" for every change shipped since that release. When a site build fails,
the first question is which Hugo, git and PostgreSQL are in use, and nothing
answers it. The license, the source and the people behind ChatCMS are not
named anywhere in the app.

## What Changes

- **An About screen** at `/administration/about`, a fourth entry in the
  platform admin area after History. Only somebody who administers the
  installation reaches it.
- **Version:** the version from `VERSION` and the short identifier of the
  build's source, such as "0.1.0 · a3bab43", or "0.1.0 · a3bab43, changed"
  for a build from an unclean tree, or "0.1.0 · development build" for a
  build that does not know it.
- **Runs on:** the versions of Hugo, git, PostgreSQL and Go the server uses,
  each found out by the server when the screen asks. One that cannot be found
  out is shown as unknown rather than failing the screen.
- **License:** Apache 2.0, with the full text readable on the screen.
- **Source:** links to the GitHub page, the changelog and the issue form.
- **Credits:** "Made by the Quiqr Team, with help from its contributors", and
  a link to the contributors page on GitHub. No names yet.
- **An answer for it:** `GET /api/administration/about`.

## Capabilities

### New Capabilities

- `platform/about`: what an installation says about itself to whoever
  administers it: version, what it runs on, license, source and credits.

### Modified Capabilities

None. The platform admin area gains an entry, which its existing requirement
already covers.

## Impact

- `flake.nix`: the deployed package passes the source identifier into the
  binary; the checks build without it (see design.md).
- `version.go` and `internal/version`: the identifier, and the license text
  embedded.
- `internal/gitsync`: a way to ask git its version, since git is called only
  from there.
- `internal/httpapi`: the about route behind `ActionAdministerInstallation`.
- The server wiring in `internal/app` for the Hugo, git and PostgreSQL
  versions.
- Frontend: a new screen, a fourth entry in `PlatformAdminShell.tsx`, strings
  in `en.ts` and `nl.ts`.
- Tests: Go unit tests for the version string and the handler, an
  integration test for the PostgreSQL version, Vitest for the screen, and an
  e2e step at both widths.
- Seams: gitsync gains one read-only call; policy is used, not changed.
- Acceptance test: none of section 6 of `docs/BRIEFING-M01.md` changes; it
  helps whoever deploys the module ("the app builds and deploys via the Nix
  flake/NixOS module") see what they deployed.

## Assumptions

- Contributors stay a link for now. Names on the screen, read from a file
  the release script writes, are a later change; the server never asks
  GitHub itself, so an offline installation shows the same screen.
- No build date: it would make every build differ, and the source identifier
  already says when.
- The list of open source packages ChatCMS is built with, and their
  notices, is not part of this change and gets a bean of its own.
- The admin area is a technical place, so the version identifier may be
  shown; the screen still avoids git words in its labels.
