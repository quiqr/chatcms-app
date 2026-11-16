---
# chatcms-g8dk
title: template gallery support
status: completed
type: feature
priority: normal
created_at: 2026-11-16T16:23:11Z
updated_at: 2026-11-16T20:00:00Z
openspec-link: openspec/changes/archive/2026-11-16-template-gallery
---

the old quiqr desktop support a public template gallery, which can be imported right away. 

## Summary of Changes

- The server reads the community template feed (or another, set with
  `services.chatcms.templateFeed`, or none), keeps the last good copy in the
  database and refreshes it hourly; screenshots are proxied and cached.
- New project gets a gallery: cards, search, loading, unavailable and
  no-match states, then a "Name & place in git" step with limit notices.
- After the key works, the template is copied into an empty remote as one
  version, "Start from <template>"; a remote that holds work is refused.
- The project records the template's Hugo version; settings show and change
  it, and say which Hugo builds it here. Builds still use the pinned Hugo.
- Design hand-off kept in `docs/design/template-gallery/`.
