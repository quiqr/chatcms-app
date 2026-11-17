---
# chatcms-tpop
title: the sidepanels need collapse icons and uncollapse icons
status: completed
type: feature
priority: normal
created_at: 2026-11-17T18:18:38Z
updated_at: 2026-11-17T23:30:00Z
openspec-link: openspec/changes/archive/2026-11-17-panels-fold-by-hand
parent: chatcms-iqog
---

I mean hide sidepanel to a icon panel and show sidepanel as full panel again.
the main sidepanel and the content sidepanel both need clear icons so they can always be shown and hidden

## Summary of Changes

Shipped with `panels-fold-by-hand` (commit 2070071):

- Only the person folds a panel now. Opening an item no longer folds the
  navigation, and both panels stay as left.
- The navigation folds to a 64 px rail of the same entries; the content menu
  folds to a 48 px strip. Each has its own fold and unfold icon.
- Both states are remembered per person, on every screen and across reloads.
- A 180 ms width motion when folding, none on load or with reduced motion,
  kept in one CSS block so it can be removed whole.
