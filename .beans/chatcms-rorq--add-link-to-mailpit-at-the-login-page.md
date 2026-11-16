---
# chatcms-rorq
title: add link to mailpit at the login page
status: completed
type: feature
priority: normal
created_at: 2026-11-16T11:26:17Z
updated_at: 2026-11-16T16:00:00Z
openspec-link: openspec/changes/archive/2026-11-16-dev-mail-link
parent: chatcms-27ng
---

- only in develment mode

## Summary of Changes

- Under `npm run dev` the sign-in and forgot-password screens link to the
  instance's Mailpit inbox, in a new tab, at the page's own host.
- `scripts/dev-fe.sh` passes the instance's inbox port; plain `npm run dev`
  falls back to 8025.
- A production build carries neither link.
