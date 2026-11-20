---
# chatcms-0xlf
title: Signing out sometimes lands on the sign-in screen without saying so
status: completed
type: bug
priority: low
created_at: 2026-11-18T21:41:32Z
updated_at: 2026-11-20T21:13:27Z
openspec-link: openspec/changes/archive/2026-11-20-steadier-under-load
---

`test/e2e/14-signing-out.spec.mjs`, "signing out ends the session, and says
so", failed once in the ship gate: after signing out the sign-in screen
showed, but not "You have signed out." Run on its own it passed 50 of 50
times.

A likely cause: since opening a project starts its workspace, a request can
still be in flight when the session ends; its 401 takes the person to the
sign-in screen by the route that does not carry the message, racing the one
that does.

Expected: signing out always says so, however many requests are still out.

## Summary of Changes

- Signing out stops refused requests (such as the workspace symbol's poll)
  from sending the person to the sign-in screen by the route without the
  message; signing in restores it. e2e 14 passed every run at 12 workers.
