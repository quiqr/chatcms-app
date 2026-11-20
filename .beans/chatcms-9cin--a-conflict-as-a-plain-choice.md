---
# chatcms-9cin
title: A conflict as a plain choice
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-20T11:45:17Z
parent: chatcms-9iuz
blocked_by:
    - chatcms-c1hm
---

When a rebase stops, the page shows "edited by two people" with keep mine or
keep theirs, per field where the model allows it. No git terminology. The
choice is made in the person's working copy and the rebase continues.

## Summary of Changes

- The same field changed by two people is a plain question with keep mine
  or keep theirs, no git words, at both widths (e2e 34). Delivered by the
  per-person workspace change under chatcms-9iuz.
