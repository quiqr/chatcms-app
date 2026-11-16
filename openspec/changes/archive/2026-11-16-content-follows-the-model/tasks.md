# Tasks

## 1. One listing

- [x] 1.1 Read `includeSubdirs` into `model.Collection`. Verify with a decode test that true, false and absent read back, and that absent means true.
- [x] 1.2 Add `model.Items(fsys, collection)` following Quiqr Desktop: subfolders unless `includeSubdirs: false`, `<name>/index.<ext>` as one bundle item, `hideIndex` hiding only `_index.*`, data folders recursive by every data extension, titles from front matter with a readable fallback, sort by `sortkey` ascending then by path. Verify with unit tests over an in-memory site for each scenario in the `content/browsing` delta, and with a new fixture site in `test/fixtures/sites/page-bundles` (bundles, subfolders, `_index.md`, a data folder) whose expected list is a golden file.
- [x] 1.3 Make `Navigate` count with `Items`, name the fallback groups "Pages" and "Collections", and warn when a single and a collection share a key. Verify with navigation tests that a count equals the length of the list, and for the fallback and the shared key.
- [x] 1.4 Make `content.Locate` take the files and find an item through `Items`, so a bundle key opens `index`, and a flat file wins over a bundle of the same name. Verify with tests for a flat file, a bundle in a subfolder, and both at once.

## 2. The menu for each person

- [x] 2.1 Add `content.Service.Menu`: the navigation with `siteDeveloper` sections kept only when `policy.Can` allows technical detail, and `disabled` sections dropped. Verify with tests for an editor, an implementer, a disabled section and `contentEditor`, and that a dropped single still opens by key.
- [x] 2.2 Fill `Browse` and `GET …/content` from the menu and `Items`. Verify with the existing content tests and an integration test against the xmin template's model, where the two "Pages" groups no longer clash and the Developer section is absent for an editor.

## 3. The API

- [x] 3.1 Add `GET …/projects/{project}/content-menu` and `GET …/projects/{project}/collections/{collection}/items` in the shape of design section "The API", with `path` only for those who may see technical detail. Verify with httpapi tests for both routes, a viewer, a missing collection, and `path` absent for an editor.
- [x] 3.2 Mark `hasChanges` and `status` from the publish review's question. Verify with an integration test: an edited item is changed, a made one is new, an untouched one is published.

## 4. Making, renaming, copying and removing

- [x] 4.1 Make `Create` write `<slug>/index.<ext>` under `content/` and `<slug>.<ext>` elsewhere, refusing when either exists. Verify with tests for both folders and both clashes, and a golden test comparing the bytes with what Quiqr Desktop writes for the same title.
- [x] 4.2 Make rename, copy and remove of a bundle act on its whole folder in one version. Verify with integration tests on a bundle holding a picture, for each scenario in the `content/pages` delta, and that going back restores the removed folder.

## 5. The content screen

- [x] 5.1 Add the strings for the fallback groups and any new labels to `en.ts` and `nl.ts`. Verify with `npm run check:strings`, `npm run check:interface` and the type check.
- [x] 5.2 Make `ContentPage.tsx` show the menu's groups from `content-menu`, singles as rows and each collection with its items from the items API. Verify with Vitest: groups in menu order, a developer section absent for an editor, bundle items by title.
- [x] 5.3 Add an e2e spec that starts from a template made of page bundles, opens a collection item, makes a new one and removes it, at both widths. Verify it runs in the `e2e` flake check.

## 6. Docs and changelog

- [x] 6.1 Describe `includeSubdirs`, page bundles, `hideIndex`, `sortkey` and `matchRole` as ChatCMS reads them in `docs/QUIQR-MODEL.md`. Verify by reading it against the specs.
- [x] 6.2 Add the entries to `CHANGELOG.md` under `[Unreleased]`, each at most two lines. Verify with `scripts/check-changelog.sh`.

## Workflow follow-up

- Ship with `scripts/ship-change.sh content-follows-the-model "<subject>"`.
- Add `openspec-link` to bean chatcms-w5ek and close it with a Summary of Changes.
