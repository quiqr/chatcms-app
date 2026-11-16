---
# chatcms-1hg1
title: Content menu and collection views (direction B)
status: completed
type: feature
priority: normal
created_at: 2026-11-16T21:51:19Z
updated_at: 2026-11-17T02:00:00Z
openspec-link: openspec/changes/archive/2026-11-17-content-menu-screens
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

## Summary of Changes

- A content column built from the model's menu sits beside the collection
  view and in the editor's slot. Groups fold and stay folded; the column folds
  away in the editor and comes back from the rail. Below 1024 px it is a
  drawer; on a phone it is a screen of its own.
- A collection view: titles, date, author, folder, thumbnail and status
  badges; search, folder tabs (a list beyond five), sort and 20 at a time,
  all done by the server and kept in the address. File names for implementers.
- A new-item dialog that checks the web address while typing and refuses a
  taken one before anything is written (`refused.item-address-taken`).
- The editor header says "Posts · 12 of 48" and steps to the previous and
  next item in the list as it was, also with Alt+Up and Alt+Down. A "Show on
  website" switch stands in for a `draft` field.
- Quick find on Ctrl+K or Cmd+K over every title the person may read.
- The old full-page content overview is gone; old editor links still open.
- Server: listings cached per site version; `around`, search, address and
  create routes.
