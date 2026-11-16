# Proposal

Bean: [chatcms-w5ek](../../../.beans/chatcms-w5ek--improve-parsing-of-the-quiqr-model.md)
(improve parsing of the Quiqr model). The screens from the design hand-off in
`docs/design/content-menu/` follow in bean
[chatcms-1hg1](../../../.beans/chatcms-1hg1--content-menu-and-collection-views-direction-b.md).

## Why

The content screen ignores the model's menu, and lists collections in a way
Quiqr Desktop does not. A template made of page bundles (`post/index.md`)
shows empty collections. A new item is written in a different shape from the
one Desktop writes. The Developer section a model keeps for site developers
is shown to editors. Before the new content screens can be built, what is
listed has to match the model and the old app.

## What Changes

- The content listing follows the model's `menu`: its sections in order,
  each with the singles and collections it names. Without a `menu` it falls
  back to "Pages" (the singles) and "Collections".
- A menu section with `matchRole: siteDeveloper` is shown only to people who
  may see technical detail. A section marked `disabled` is not shown.
- A collection lists its items the way Quiqr Desktop does:
  - it searches subfolders, unless `includeSubdirs: false`;
  - a page bundle (`<name>/index.md`) is one item, named after its folder;
  - `hideIndex` hides only `_index.*`, the list page of a section;
  - items are sorted by `sortkey` from the front matter when the model gives
    one, else by their path;
  - an item's title comes from its front matter `title`, falling back to a
    readable form of its file or folder name.
- A new item in a collection under `content/` is written as
  `<slug>/index.<extension>`, as Desktop does. Renaming, copying and removing
  such an item act on its folder, so its pictures go with it. Items that
  already exist keep the shape they have.
- The API gains the content menu and a collection's items in the shape the
  design asks for (`docs/design/content-menu/CONTENT-MENU.md` section 7). The
  current content screen shows the menu's sections with it; the new screens
  come in chatcms-1hg1.
- One listing serves the browser, the menu counts and opening an item, so a
  count can never disagree with the list.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `content/content-navigation`: menu sections reach the screen; `matchRole`
  and `disabled` are honoured; the fallback groups are named.
- `content/browsing`: a collection's items follow Quiqr Desktop's rules for
  subfolders, page bundles, `hideIndex`, sorting and titles.
- `content/pages`: a new item under `content/` is a page bundle, and renaming,
  copying and removing a bundle act on its folder.

## Impact

- `internal/model`: one item listing shared by navigation and browsing;
  `includeSubdirs` read into the schema.
- `internal/content`: browse, open, create, rename, copy and remove go through
  that listing and the item's file rather than rebuilding a path from its key.
- `internal/httpapi`: `GET …/content-menu` and
  `GET …/collections/{collection}/items`; `GET …/content` keeps working.
- `internal/policy` is asked who may see a developer's section; nothing else
  about roles moves.
- `web/src/routes/ContentPage.tsx` shows the menu's sections.
- Golden tests: a page made in a collection compares against what Desktop
  writes for the same title.

## Acceptance test served

Acceptance test 4 (create, edit and delete a page) on a template made of page
bundles, and test 3 (all forms from the model render), which a collection
that lists nothing cannot pass.

## Assumptions

- `matchRole` values other than `siteDeveloper` (such as `contentEditor`) show
  the section to everybody, as Desktop shows them to its default user.
- Items sort ascending by `sortkey`; Desktop leaves the direction to a toggle
  on the screen, which the collection view in chatcms-1hg1 brings.
- An item's key is the collection key and its path without the extension, with
  a bundle named by its folder (`posts/2024/spring-sale`). Existing keys of
  flat files do not change, so open links keep working.
- A single and a collection with the same key are reported as a model
  warning; the collection keeps the key in the content screen.
- A collection outside `content/` (a data folder) lists every format
  extension recursively and is not turned into bundles, as in Desktop.
