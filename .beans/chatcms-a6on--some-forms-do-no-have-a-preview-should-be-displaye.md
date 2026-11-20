---
# chatcms-a6on
title: some forms do no have a preview. Should be displayed
status: completed
type: feature
priority: normal
created_at: 2026-11-20T11:45:27Z
updated_at: 2026-11-20T14:21:57Z
openspec-link: openspec/changes/archive/2026-11-20-preview-without-a-page
---

There are more possibilies:

- One site with multiple data form. E.g. UIlite. This is a one pages but has many data items. The preview should be always the same I think. Maybe some info so the user understands.

- A site with some data forms tight to a preview page, and some data forms having global data or anything else... I think every item that does not have it own page should not show a preview. The preview should default to "no preview available" or something more informative...

## Summary of Changes

- Agreed on 20 November 2026: an item without a page of its own gets an
  explanation in the preview, with "Show the home page".
- An item is data when its file is JSON, YAML or TOML; the model's
  `previewUrl` (show that address) and `hidePreviewIcon` (no preview) win.
- A one-page site such as UILite can set `previewUrl: /` on its data singles
  to show the home page always.
