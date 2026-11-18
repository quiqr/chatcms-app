---
# chatcms-8g3i
title: A working copy per person per project
status: todo
type: feature
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T12:51:48Z
parent: chatcms-gkdd
blocked_by:
    - chatcms-7jbk
---

Each (user, project) gets its own working copy on its own branch, created from
the shared version on first open. Saving writes and auto-commits to that
person's copy. `gitsync` gains this in its own vocabulary, without branch
names in its interface.
