---
# chatcms-9m9u
title: after a change, the publish counter doesn't update
status: completed
type: bug
priority: normal
created_at: 2026-11-20T08:39:24Z
updated_at: 2026-11-20T12:32:36Z
openspec-link: openspec/changes/archive/2026-11-20-publish-count-follows-saves
parent: chatcms-iqog
---

## Summary of Changes

- Every successful write to a project, and every assistant change, makes
  the screens read the project again (at once and two seconds later), so the
  Publish count follows without a reload. Proven by e2e 50 at both widths.
