# Design

## Context

Two pieces of code answer "what is in this collection", and they disagree.
`model.Navigate` follows the menu and counts files filtered by extension, but
nothing outside its tests calls it. `content.Browse` builds "Pages" and then
every collection, lists only files directly in the folder, skips folders, and
hides `index.*` as well as `_index.*` when `hideIndex` is set. `content.Locate`
turns a key back into a file by appending the extension, so it can never find
a page bundle. New items are written as `<slug>.<ext>`.

Quiqr Desktop (`packages/backend/src/services/workspace/workspace-service.ts`
in `../quiqr-desktop`) is the reference:

- `listCollectionItems` globs `**/*.{md,html,markdown,qmd}` under a folder in
  `content/`, or `*` when `includeSubdirs` is false.
- `hideIndex` turns the glob into `!(_index).*`.
- A match ending in `/index.<ext>` is a page bundle, labelled by its folder.
- Other folders glob every data extension recursively.
- `sortkey` reads a front-matter value.
- `createCollectionItemKey` writes `<key>/index.<ext>` under `content/`.
- Rename and delete act on a bundle's folder.

## Goals / Non-Goals

**Goals:**

- One listing of a collection's items, used by navigation, browsing, the
  counts and opening an item.
- The same items, keys and new files as Quiqr Desktop, so a site moves between
  the two without surprises.
- The API the new content screens need (design section 7), served now.

**Non-Goals:**

- The new screens: content column, collection view, quick find, previous and
  next, phone screens. They are chatcms-1hg1.
- Folder tabs, searching and paging on the server. The items API returns the
  whole list with each item's folder; chatcms-1hg1 adds the parameters.
- Turning an existing flat file into a page bundle, which Desktop offers as
  its own action.

## Decisions

### One listing in the model package

`model.Items(fsys, collection) []Item` walks the collection's folder by
Desktop's rules and returns, per item:

- its key within the collection;
- its file;
- whether it is a bundle;
- its folder below the collection;
- its sort value.

The title is read with the front matter, so the listing reads each file once.
`Navigate` counts with it, `Browse` lists with it, and `Locate` finds an
item's file by its key in it.

Alternative: keep the two and make them agree by tests. They disagree today
despite tests, because each was tested against its own idea of a collection.

### Keys stay as they are for flat files

An item's key is `<collection>/<path without extension>`, and a bundle's path
is its folder. `posts/hello` keeps meaning `content/post/hello.md`, so links
and history entries already made keep working. When `hello.md` and
`hello/index.md` both exist, the flat file wins and the model warnings say so.

### `Locate` needs the files

`Locate(m, key)` becomes `Locate(m, fsys, key)`, because a key no longer says
which file it names. Every caller already has the working copy open.

### The menu reaches the screen through the content service

`Service.Menu` returns the navigation for the actor. It drops a
`siteDeveloper` section unless `policy.Can(actor, ActionSeeTechnicalDetail,
project)`, and drops a `disabled` one for everybody. The decision is the
policy's; the content package only asks. Singles in a dropped section stay
reachable by key, because hiding is about the menu, not about access.

### The API

- `GET …/projects/{project}/content-menu` returns groups of entries, each
  with:
  - `key`;
  - `kind` (`single` or `collection`);
  - `title`;
  - `count`, for a collection;
  - `hasChanges`.

  It also returns totals and the model warnings.
- `GET …/projects/{project}/collections/{collection}/items` returns every item
  with:
  - `key`, `title` and `folder`;
  - `status` (changed, new, or published);
  - `path`, only for those who may see technical detail.

  It also returns the folders with their counts.
- `GET …/content` keeps its shape, filled from the same listing, until the
  content screen moves to the two above.

`hasChanges` and `status` come from the same question the publish review
asks, as the browsing spec requires.

### New items under `content/` are bundles

`Create` writes `<folder>/<slug>/index.<ext>` for a collection under
`content/` and `<folder>/<slug>.<ext>` elsewhere, refusing when either
`<slug>.<ext>` or `<slug>/` exists. Rename and copy of a bundle move or copy
every file under its folder in one version; remove deletes them in one
version. A golden test compares a made page with what Desktop writes.

### The current content screen shows the menu

`ContentPage.tsx` reads `content-menu` and shows each group with its entries:
a single opens its form, and a collection shows its items as it does now,
read from the items API. This is a stopgap that the collection view replaces
in chatcms-1hg1.

## Risks / Trade-offs

- [Reading every item's front matter on each listing is slow for a large
  collection] → the listing reads only the front matter block, and a site of
  a few hundred items stays well under the screen's budget. A cache keyed by
  the version can come with paging in chatcms-1hg1.
- [A site that relied on `index.md` being hidden by `hideIndex`] → that
  differs from Desktop. Desktop shows such files, so ChatCMS now does too.
- [Bundle folders hold pictures that removing deletes] → the confirmation
  already says the page can be brought back, and going back restores the
  whole folder because it is one version.
