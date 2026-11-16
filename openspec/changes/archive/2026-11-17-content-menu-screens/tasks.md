# Tasks

## 1. The server

- [x] 1.1 Cache the collection listing per site version. Verify with a test that a second listing at the same version reads no files, and that a save makes the next one read again.
- [x] 1.2 Give the items route `q`, `folder`, `sort`, `page`, `pageSize` and `around`, and the folders with counts. Verify with httpapi and integration tests for each sort, a search, a folder, paging, `around` at both ends, and the whole list when nothing is asked.
- [x] 1.3 Add `draft` to an item's status from its front matter. Verify with a test for an item with `draft: true`.
- [x] 1.4 Add `POST …/collections/{collection}/items` and `GET …/collections/{collection}/address`, with `refused.item-address-taken` and the file only for those who may see technical detail. Verify with tests: a free address, a taken one, a folder, `previewUrlBase`, and no file for an editor.
- [x] 1.5 Add `GET …/search?q=` over singles and items the person may read. Verify with tests: grouped results, the limit of 20, and a developer's single absent for an editor.

## 2. The content column

- [x] 2.1 Add the strings for every screen in this change to `en.ts` and `nl.ts`. Verify with `npm run check:strings`, `npm run check:interface` and the type check.
- [x] 2.2 Build `ContentLayout` and the content column: groups, rows, counts, dots, the current row, folding groups, remembered per person and project. Verify with Vitest for each scenario in the `content/content-menu` delta about the column and its groups.
- [x] 2.3 Route `/content`, `/content/c/:collection` and `/content/*` through the layout, keeping the editor's address. Verify with Vitest that an old editor link still opens and `/content` opens the first view.

## 3. The collection view

- [x] 3.1 Build the list: heading, rows with title, meta, thumbnail and badge, paging, empty and no-results states, with the query in the address. Verify with Vitest for each list scenario in the `content/collection-view` delta.
- [x] 3.2 Add the search with a 250 ms pause, the folder tabs or list, the sort and the file names switch for implementers. Verify with Vitest: the server is asked once per pause, tabs hide without folders, a list appears beyond five, and the switch is absent for an editor.
- [x] 3.3 Add the row actions and the new-item dialog with the address check. Verify with Vitest: Create and edit opens the editor, a taken address keeps it unavailable, a viewer gets no actions.

## 4. The editor

- [x] 4.1 Put the content column in the editor's slot, with the fold button and the rail's Content icon, remembered. Verify with Vitest that the fold survives a remount.
- [x] 4.2 Add the header: back, "<collection> · n of N", previous and next from `around`, Alt+Up and Alt+Down ignored while typing. Verify with Vitest for each scenario in the `content/editing` delta about stepping.
- [x] 4.3 Add "Show on website" for a `draft` field. Verify with Vitest that it is inverted and absent without the field, and with the golden test that the saved front matter matches the form's boolean.

## 5. Quick find and narrow screens

- [x] 5.1 Add quick find on ⌘K or Ctrl+K. Verify with Vitest: it opens from any project screen, groups results, Enter opens the first, nothing found is said.
- [x] 5.2 Make the content column a drawer below 1024 px and three screens in the phone frame below 640 px. Verify with Vitest at both sizes that choosing closes the drawer and the phone screens follow menu, collection, editor.
- [x] 5.3 Remove the old content overview and its strings. Verify with `npm run check:strings` that no unused key remains, and the type check.

## 6. End to end

- [x] 6.1 Add an e2e spec covering the design's acceptance checks 1 to 8 at 1440 and 390 px, on a template with a menu, subfolders and page bundles. Verify it runs in the `e2e` flake check.
- [x] 6.2 Update the specs that click through the old content overview. Verify the whole `e2e` check passes.

## 7. Docs and changelog

- [x] 7.1 Point `docs/design/DESIGN.md` to `docs/design/content-menu/` for the content area. Verify by reading it.
- [x] 7.2 Add the entries to `CHANGELOG.md` under `[Unreleased]`, each at most two lines. Verify with `scripts/check-changelog.sh`.

## Workflow follow-up

- Ship with `scripts/ship-change.sh content-menu-screens "<subject>"` after `content-follows-the-model`.
- Add `openspec-link` to bean chatcms-1hg1 and close it with a Summary of Changes.
