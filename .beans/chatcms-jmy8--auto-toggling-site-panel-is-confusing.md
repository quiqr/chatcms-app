---
# chatcms-jmy8
title: auto toggling site panel is confusing
status: completed
type: feature
priority: normal
created_at: 2026-11-17T19:01:43Z
updated_at: 2026-11-17T23:30:00Z
openspec-link: openspec/changes/archive/2026-11-17-panels-fold-by-hand
parent: chatcms-iqog
---

I don't know what the solution is but this is confusion and disturbing

## Summary of Changes

Shipped with `panels-fold-by-hand` (commit 2070071):

- Only the person folds a panel now. Opening an item no longer folds the
  navigation, and both panels stay as left.
- The navigation folds to a 64 px rail of the same entries; the content menu
  folds to a 48 px strip. Each has its own fold and unfold icon.
- Both states are remembered per person, on every screen and across reloads.
- A 180 ms width motion when folding, none on load or with reduced motion,
  kept in one CSS block so it can be removed whole.
