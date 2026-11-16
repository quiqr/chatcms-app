# Tasks

## 1. Keeping a site

- [x] 1.1 Add a goose migration for `projects.hosted` and queries to set and read it. Verify with an integration test that it reads back.
- [x] 1.2 Add `Sites.Host(ctx, id)` making `<state>/hosted/<id>.git` as a bare repository and returning its path. Verify with a gitsync test that the repository exists, is bare and is empty.
- [x] 1.3 Let a template or export import with no address use `Host` as its remote and set `hosted`; refuse a repository import with no address. Verify with integration tests: a template without an address is ready with one version in the kept repository and no key, an export without an address likewise, a repository import without one is refused.

## 2. Connecting later

- [x] 2.1 Add `Repo.MoveTo(ctx, remote)` that sets `origin` and sends the main line, putting `origin` back on failure. Verify with a gitsync test: the new remote holds every version, and an unreachable one leaves `origin` unchanged.
- [x] 2.2 Add `ConnectSite` to the onboarding service and `PUT …/projects/{project}/remote`: permission, kept, key reaches, empty, move, record, clear `hosted`, remove the kept repository. Verify with integration tests for each scenario in the `tenancy/projects` delta, and an httpapi test for the route.
- [x] 2.3 Remove the kept repository when a kept project is removed. Verify with an integration test.

## 3. Publishing

- [x] 3.1 Refuse publishing a kept site with `refused.site-not-connected`, before any publish record. Verify with an integration test that the refusal carries the key and no record exists, and that a connected site publishes.

## 4. The screens

- [x] 4.1 Add the strings to `en.ts` and `nl.ts`. Verify with `npm run check:strings`, `npm run check:interface` and the type check.
- [x] 4.2 Make the git address optional on the "Name & place in git" step, with the hint, and let the setup page copy a kept template without a key. Verify with Vitest: continuing with no address, and the copy enabled without a key.
- [x] 4.3 Let the setup page bring in a Quiqr Desktop export without an address. Verify with Vitest that the import is asked for with an empty remote.
- [x] 4.4 Show "Kept in ChatCMS, not connected yet" in project settings with the key and an address to connect, and say a publish refusal plainly on the publish screen. Verify with Vitest for connecting, a refusal, and the publish message.
- [x] 4.5 Extend `test/e2e/29-start-from-a-template.spec.mjs`: start from the template with no address, edit, then connect an empty remote in settings and find the versions there, at both widths. Verify it runs in the `e2e` flake check.

## 5. Docs and changelog

- [x] 5.1 Add `<state>/hosted/` to "What to back up" in `docs/DEPLOYING.md`, and note the change of assumption 1 in `docs/design/template-gallery/TEMPLATE-GALLERY.md`. Verify by reading both.
- [x] 5.2 Add the entries to `CHANGELOG.md` under `[Unreleased]`, each at most two lines. Verify with `scripts/check-changelog.sh`.

## Workflow follow-up

- Ship with `scripts/ship-change.sh hosted-until-connected "<subject>"`.
- Add `openspec-link` to bean chatcms-x2uz and close it with a Summary of Changes.
