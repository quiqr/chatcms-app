---
# chatcms-ifc4
title: 'whats means: sign our everywhere?'
status: completed
type: task
priority: normal
created_at: 2026-11-17T18:21:11Z
updated_at: 2026-11-18T01:00:00Z
openspec-link: openspec/changes/archive/2026-11-17-sign-out-only
---

## Summary of Changes

Shipped with `sign-out-only` (commit d8e1733):

- "Sign out everywhere" ended every session of the account on every device,
  including the one in use. It is removed from the person menu, which now
  offers only "Sign out".
- The endpoint and service method behind it are removed; password reset and
  account removal still end every session.
- The account page keeps its list of where somebody is signed in, with an End
  button per device.
