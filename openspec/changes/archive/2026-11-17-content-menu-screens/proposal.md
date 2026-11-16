# Proposal

Bean: [chatcms-1hg1](../../../.beans/chatcms-1hg1--content-menu-and-collection-views-direction-b.md)
(content menu and collection views, direction B). It builds on
`content-follows-the-model` (bean chatcms-w5ek), which must ship first.

## Why

Quiqr Desktop lets a content creator pick a collection, see its items and
open one. ChatCMS shows one long page holding every single and every item,
which stops working at a few dozen posts and loses the structure the
developer gave the menu. The design hand-off in `docs/design/content-menu/`
(direction B) is the chosen replacement.

## What Changes

- **Content column.** Inside a project the content area gets a column built
  from the model's menu:
  - groups in model order, each collapsing, with the state remembered per
    person and project;
  - singles as page rows, collections as rows with their item count;
  - a dot where something waits to be published;
  - the open item's collection stays highlighted.
- **Collection view.** Picking a collection shows its items:
  - title, date, author and folder, each where the model has one;
  - a status badge per item, and a thumbnail when the model has an image
    field;
  - search within the collection, folder tabs when it has subfolders, and
    sort (newest, oldest, title, recently edited);
  - "Show 20 more" for paging;
  - empty and no-results states.
- **File names.** Implementers and admins can switch on each item's file
  name. Editors never see a path.
- **New item dialog.** It asks for a title, and a folder when there are
  subfolders, and previews the web address, which can be edited. Implementers
  also see the file that will be made. An address that is taken is refused
  before anything is written.
- **Row actions.** Open, rename, duplicate and remove, from each row of the
  collection view. Moving to another folder is left out; see Assumptions.
- **Editor header.** With an item open:
  - back to its collection, and "<collection> · 12 of 48";
  - previous and next through the current sort, search and folder,
    with Alt+Up and Alt+Down as keys;
  - a "Show on website" switch when the model has a `draft` field.
- **Folding.** In the editor the app sidebar becomes the 64 px rail. The
  content column folds away and comes back from the rail's Content icon, and
  the choice is remembered.
- **Quick find.** ⌘K or Ctrl+K anywhere in a project finds singles and
  collection items by title, grouped, and Enter opens one.
- **Narrower screens.** Below 1024 px the content column is a drawer. Below
  640 px content is three screens (menu, collection, editor) inside the
  phone frame's existing tab bar.
- **API.** The items API gains search, folder, sort and paging. Making an item
  answers a taken address with its own refusal. A search endpoint serves
  quick find.
- The old full-page content overview is removed.

## Capabilities

### New Capabilities

- `content/content-menu`: the content column, its groups, folding, quick
  find, and the narrow-screen and phone forms of it.
- `content/collection-view`: a collection's item list with search, folders,
  sort, paging, status, file names and the new-item dialog.

### Modified Capabilities

- `content/editing`: the editor header knows the item's place in its
  collection, steps to the previous and next, and offers "Show on website"
  for a `draft` field.
- `interface/app-shell`: the editor's empty slot becomes the content column,
  which folds away.

## Impact

- `web/src/routes/ContentPage.tsx` is replaced by a content layout with the
  content column, a collection view and the editor. `EditorPage.tsx` gains the
  header.
- New components: content column, collection list, new-item dialog, quick
  find, phone menu and collection screens.
- `internal/content` and `internal/httpapi`: query parameters on the items
  route, a refusal key for a taken address, and `GET …/search`.
- `internal/policy` is asked who may see file names; nothing else moves.
- Strings in `en.ts` and `nl.ts`; e2e specs at 1440 and 390 px.

## Acceptance test served

Acceptance test 4 (create, edit and delete a page) and test 9 (everything
works on a phone), for sites with more than a handful of items.

## Assumptions

- "Move to folder" from the design's row menu is left out. Renaming within a
  folder covers what Desktop offers, and moving a bundle between language
  folders needs a decision about translations first.
- "Recently edited" sorts by the time of the last version that touched the
  item, from history, not by the file's time on disk.
- The thumbnail is the first image field the model gives the collection. An
  item without one shows no thumbnail rather than a placeholder.
- An item's date and author come from front-matter fields called `date` and
  `author` when the model has them.
- Fold and collapse state live in the browser's local storage, as the design
  allows.
- The "Show on website" switch writes the `draft` field the same way a
  boolean field in the form does, so the golden tests still hold.
