# Proposal

Bean: [chatcms-ifc4](../../../.beans/chatcms-ifc4--whats-means-sign-our-everywhere.md)
(what does "Sign out everywhere" mean?).

## Why

The person menu offers "Sign out" and, right under it, "Sign out everywhere".
Nobody can tell from the words what the second one does or when to use it. It
ends every session the account holds, on every device, which is almost never
what somebody wants: being signed in on another device is rare, and whoever
is signed in there can sign out there.

## What Changes

- **BREAKING:** "Sign out everywhere" is removed from the person menu. The
  menu offers only "Sign out".
- **BREAKING:** `POST /api/sign-out/everywhere` is removed, with the service
  method behind it. Its only caller was the menu item.
- Unchanged: ending one session from "Where you are signed in" on the
  account page, and the password reset and account removal that end every
  session of an account.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `identity/authentication`: the requirement "Somebody can sign out
  everywhere" is removed.

## Impact

- Frontend: `web/src/shells/PersonMenu.tsx`, `web/src/lib/session.ts`, the
  string in `en.ts` and `nl.ts`, and the assertion in `shells.test.tsx`.
- Backend: the route in `internal/httpapi/router.go`, its handler and the
  interface method in `internal/httpapi/auth.go`, `SignOutEverywhere` in
  `internal/auth/service.go`, and their tests. `Sessions.RevokeAll` stays,
  because password reset and account removal use it. The auth seam loses one
  method; nothing else behind it changes.
- e2e: the "signing out everywhere" test in `14-signing-out.spec.mjs` goes.
- Acceptance test: none of section 6 of `docs/BRIEFING-M01.md` depends on it.

## Assumptions

- Nobody relies on the endpoint outside this frontend: it is not documented
  as an API and the app is not yet released beyond 0.1.0.
- The account page's list of where somebody is signed in, with its End
  button per device, stays. It answers the rare case without a menu item.
