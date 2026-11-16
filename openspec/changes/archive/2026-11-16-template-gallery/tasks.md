# Tasks

## 1. The feed

- [x] 1.1 Record the live feed as `test/fixtures/templates/templates.json`, with one entry broken by hand. Verify the fixture holds nine valid entries and the broken one.
- [x] 1.2 Map feed entries to templates in `internal/onboarding`: drop invalid ones, split the Hugo version and the extended flag, split the content model, sort by name. Verify with unit tests on the fixture: eight kept, alphabetical, `extended_0.101.0` becomes `0.101.0` extended.
- [x] 1.3 Add a goose migration for `template_feed` and sqlc queries to store and read the last good copy. Verify with an integration test that a stored copy is read back unchanged.
- [x] 1.4 Add the `templates.refresh` job: fetch, validate, store, enqueue the next run an hour later; leave the stored copy alone on failure. Verify with integration tests against an `httptest` server: a good fetch stores, a failing one keeps the earlier copy, and the next run is enqueued.
- [x] 1.5 Add `services.chatcms.templateFeed` to the NixOS module, passed as `CHATCMS_TEMPLATE_FEED`, `null` switching it off; the server enqueues the first refresh at start-up only when set. Verify with the module's VM test that the default reaches the service, and a unit test that an empty setting enqueues nothing.

## 2. The API

- [x] 2.1 `GET /api/templates` returns the stored list, or a refusal with its key when none was ever fetched or the gallery is off; only for people who may create projects. Verify with httpapi tests for each case and for an editor being refused.
- [x] 2.2 `GET /api/templates/{id}/screenshot` fetches, type-checks by bytes, caches on disk and serves the screenshot with a long cache lifetime. Verify with tests: first request fetches, second does not, a non-image is refused, an unknown id is not found.
- [x] 2.3 Read the installation's Hugo version at start-up and return it beside the template list. Verify with a unit test parsing `hugo version` output.

## 3. Starting from a template

- [x] 3.1 Add `hugo_version` and `hugo_extended` to projects in a migration, with queries to set and read them. Verify with an integration test.
- [x] 3.2 Replace the template import: look the template up in the stored list, refuse a non-empty remote, clone `SourceLink`, drop its history, `sites.Create` with "Start from <name>", record the Hugo version. Verify with integration tests against on-disk remotes: an empty remote gets one version with the template's files, a non-empty one is refused and untouched, an unknown template is refused.
- [x] 3.3 Remove the empty `Templates` map and the typed template name. Verify that `go vet ./...` and the frontend type check pass.
- [x] 3.4 Show and change the Hugo version in project settings, with the note when it differs from the installation's. Verify with Vitest on the settings screen.

## 4. The screens

- [x] 4.1 Add strings for the gallery, the step and the settings note to `en.ts` and `nl.ts`. Verify with `npm run check:strings` and the type check.
- [x] 4.2 Build the gallery as the third source of a new project, following `docs/design/template-gallery/reference/`: cards, search with count, live demo, loading, error with try again, no match, footer link. Verify with Vitest for search, count, no match, error, no demo link, and an editor not seeing it.
- [x] 4.3 Build the step "Name & repository": template details, project name, organisation with limit notice, empty repository, Hugo note, then the existing deploy key, publishing and invite steps. Verify with Vitest for a taken name, the last free project, a full organisation, and the Hugo note.
- [x] 4.4 Make both screens work at 390 px as in `reference/screenshots/phone/`. Verify with `npm run check:interface` and the e2e check below.
- [x] 4.5 Add a Playwright test that starts a project from a template served by a local feed into an empty on-disk remote, at 1440 and 390 px. Verify it runs in the `e2e` flake check.

## 5. Docs and changelog

- [x] 5.1 Document `templateFeed` in `docs/DEPLOYING.md`, and add §5.3a to `docs/design/DESIGN.md` pointing at `docs/design/template-gallery/`. Verify by reading both.
- [x] 5.2 Add the gallery to `CHANGELOG.md` under `[Unreleased]`, each entry at most two lines. Verify with `scripts/check-changelog.sh`.

## Workflow follow-up

- Ship with `scripts/ship-change.sh template-gallery "<subject>"`.
- Add `openspec-link` to bean chatcms-g8dk and close it with a Summary of Changes.
- Open a bean for building with a project's own Hugo version once microVMs exist.
