---
# chatcms-yr88
title: Background rebase of each person's changes
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-19T10:07:29Z
openspec-link: openspec/changes/archive/2026-11-18-add-workspace-per-person
parent: chatcms-9iuz
---

After each auto-commit, the person's branch is rebased onto the shared version
and the result is pushed to it, in a job. A clean rebase needs nothing from
the user.

## Summary of Changes

Delivered by `openspec/changes/archive/2026-11-18-add-workspace-per-person`: every save is shared with the shared version in the
background, by merging rather than rebasing (the open decision the epic
flagged). A clean merge needs nothing from anyone; the same field changed
twice becomes a question. `openspec/changes/archive/2026-11-19-add-ai-combine` added that a copy with nothing of its own
follows the shared version when an answer rewrites its history.
