---
# chatcms-6a72
title: per page undo with a visual undo history
status: completed
type: feature
priority: normal
created_at: 2026-11-17T19:12:59Z
updated_at: 2026-11-18T00:30:00Z
openspec-link: openspec/changes/archive/2026-11-17-page-undo
---

Every change on a page is saved at once, 800 ms after typing stops. That is
safe against losing work, but it leaves no quick way back from a mistake: a
cleared field or a deleted paragraph is saved before anybody notices.

The editor needs an undo for the page that is open:

- Undo and redo buttons in the editor header, and Ctrl+Z / Cmd+Z outside a
  text field. Inside a text field the browser's own undo keeps working.
- A visual undo history: a list of the steps since the page was opened, each
  named after the fields it changed, with its time. Picking a row goes back
  to that point.
- An undo is a save like any other, so the version history keeps everything
  and nothing can be lost by undoing.
- Undo puts back only the fields of that step, so a field somebody else
  changed in the meantime is left alone.
- No backend work: the steps live in the browser and survive a reload of the
  tab.

Going back to a version of one page from days ago (page history) is a
different need and is not part of this bean.

## Summary of Changes

Shipped with `page-undo` (commit 5931eda):

- Undo, Redo and Undo history buttons beside the saved status in the editor,
  at every width; Ctrl+Z or Cmd+Z and Ctrl+Shift+Z or Ctrl+Y outside a field.
- One save is one step. An undo is a save like any other, so nothing is lost,
  and it leaves alone a field that changed since.
- The undo history lists the steps since the page was opened; picking a row
  goes back to that point in one save.
- Steps are kept for the browser tab and survive a reload, up to 100.
- Going back to older versions of one page stays out of scope.
