# Proposal

Bean: [chatcms-rorq](../../../.beans/chatcms-rorq--add-link-to-mailpit-at-the-login-page.md)
(add link to mailpit at the login page), milestone chatcms-27ng (05 develop
happiness).

## Why

`scripts/dev-be.sh` catches every mail an instance sends in Mailpit, but the
only place its address appears is the terminal notice. On a development
machine, the sign-in and forgot-password screens are where you wait for an
invitation or a reset mail, so the inbox should be one click away from there.

## What Changes

- Under the frontend development server, the sign-in screen's development
  block and a new development block on the forgot-password screen link to
  the Mailpit inbox of the instance.
- `scripts/dev-fe.sh` passes the instance's inbox port on to the frontend.
  The link takes its host from the page address, so it also works from a
  phone.
- A production build carries neither block.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `platform/local-development`: adds a requirement for the inbox link on the
  sign-in and forgot-password screens in development.

## Impact

- `web/src/routes/DevMailLink.tsx` (new), `DevSignIn.tsx`,
  `ForgotPasswordPage.tsx`, `web/vite.config.ts`, `web/src/strings/en.ts` and
  `nl.ts`.
- `scripts/dev-fe.sh` and its test.
- No server code changes. No seam is touched.

## Acceptance test served

None from `docs/BRIEFING-M01.md` section 6 directly. It shortens the manual
path through the invitation and reset tests on a development machine.

## Assumptions

- Mailpit's inbox port is 8025, or 8026 for a named instance, as
  `scripts/dev-instance.sh` sets it. Started with plain `npm run dev`, the
  frontend assumes 8025.
- The link is shown even when Mailpit is not running. Checking first is not
  worth a request from the browser.
- The link opens in a new tab, so the screen with the form is kept.
