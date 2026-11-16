---
# chatcms-w5ek
title: improve parsing of the quiqr menu's
status: completed
type: feature
priority: normal
created_at: 2026-11-16T16:05:41Z
updated_at: 2026-11-17T00:30:00Z
openspec-link: openspec/changes/archive/2026-11-16-content-follows-the-model
---

This might be to big but what could be achieved right now:

1. Quiqr Desktop has menu definition in which is configured which single or
   collection is opened. The developer has influence how the content menu is
   structured. Is this implemented?

2. Collections are not working. In Quiqr Desktop you get a list of items, e.g.
   markdown files then you choose one and get the form.

3. In hugo index.md and _index.md have special meaning. The quiqr model
   proporties for handling index.md's

There are more to come. Lets starts with these.

There is a design zip here: /home/pim/Downloads/quiqr-chatcms-content-menu.zip

## Summary of Changes

- The content screen follows the model's `menu`, in order; without one it
  shows "Pages" and "Collections". A `matchRole: siteDeveloper` section is
  only for those who see technical detail, and a `disabled` one for nobody.
- Collections list items by Quiqr Desktop's rules: subfolders unless
  `includeSubdirs: false`, `<name>/index.md` as one page bundle, `hideIndex`
  hiding only `_index.*`, `sortkey` ordering, titles from front matter.
- A new item under `content/` is written as `<slug>/index.md` with the bytes
  Desktop writes; renaming, copying and removing a bundle act on its folder.
- New API: `GET …/content-menu` and `GET …/collections/{collection}/items`,
  with file paths only for implementers.
- A collection called `pages` no longer clashes with the Pages group.
- The screens of the design hand-off follow in chatcms-1hg1.
