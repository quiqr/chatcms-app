---
# chatcms-1hg1
title: Content menu and collection views (direction B)
status: in-progress
type: feature
priority: normal
created_at: 2026-11-16T21:51:19Z
updated_at: 2026-11-16T21:59:40Z
---

The second half of chatcms-w5ek: the content menu screens from the design
hand-off in `docs/design/content-menu/` (direction B).

- A content column built from the model's menu, that folds next to the editor.
- A collection view: folder tabs, sort, search, paging, status badges.
- A new-item dialog with the web address and, for implementers, the file path.
- Previous and next in the editor, and quick find with Cmd/Ctrl+K.
- The phone screens: menu, collection, editor.

It builds on the change `content-follows-the-model`, which makes the content
listing follow the Quiqr model and adds the content menu and collection item
API this needs.
