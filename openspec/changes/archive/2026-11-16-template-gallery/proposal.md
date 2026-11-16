# Proposal

Bean: [chatcms-g8dk](../../../.beans/chatcms-g8dk--template-gallery-support.md)
(template gallery support).

Design hand-off: `docs/design/template-gallery/` (spec in
`TEMPLATE-GALLERY.md`, reference screens at 1440 and 390 px).

## Why

The old Quiqr offers a public gallery of community templates that can be
started from at once. ChatCMS has a "Start from a template" option, but it
asks for a typed name and knows no templates, so it cannot be used. The
briefing's onboarding section promises starting from a Quiqr template.

## What Changes

- The server reads a template feed, by default the Quiqr community feed that
  the old Quiqr uses. A job fetches and validates it, keeps the last good
  copy, and the server proxies and caches the screenshots. An installation can
  point at another feed, or switch the gallery off.
- Creating a project gets a gallery as the third source, "Start from a
  template": cards with screenshot, description, author and licence, live
  demo, search, and loading, error and no-match states.
- A second step, "Name & repository", shows the template's details and asks
  for a project name, the organisation and an empty repository. The existing
  onboarding follows: deploy key, publishing, invite people.
- After the deploy key works, the template is copied into the empty
  repository as one fresh version, "Start from <template>", and sent there.
  A repository that is not empty is refused.
- The project records the template's Hugo version. Project settings show it
  and allow changing it.
- The typed template name field and the empty list behind it go away.

## Capabilities

### New Capabilities

- `onboarding/template-gallery`: the feed, its cache, the screenshots, the
  gallery and choosing a template.

### Modified Capabilities

- `onboarding/import`: starting from a template copies it into an empty
  repository with a fresh history, rather than cloning it as the site.
- `tenancy/projects`: a project records the Hugo version it asks for.
- `platform/deployment`: the module sets the template feed, or switches it
  off.

## Impact

- `internal/onboarding` (feed, cache, template import), `internal/jobs`
  (refresh job), `internal/httpapi` (template list and screenshot routes),
  a migration for the project's Hugo version, the NixOS module.
- `web/src/routes/OnboardingPage.tsx` and new gallery and step screens,
  strings in `en.ts` and `nl.ts`.
- The `sync/repository` seam is used, not changed: the fresh copy goes
  through the same create-and-send path as the export import.

## Acceptance test served

None of the nine in `docs/BRIEFING-M01.md` section 6 directly. It fulfils the
briefing's onboarding promise, "start from a Quiqr template", and the hand-off's
own acceptance checks in `TEMPLATE-GALLERY.md` section 7.

## Assumptions

- **Hugo version:** stored on the project now and editable in settings, but
  builds keep using the Hugo the installation pins. A per-project Hugo comes
  with the planned microVMs. Until then the step and the settings say when
  the template asks for a different version than the installation runs. This
  departs from the hand-off's acceptance check 6, which expects the build to
  use the template's version.
- **API:** the hand-off sketches `POST /api/orgs/{org}/projects/from-template`.
  ChatCMS already creates a project and imports into it with two calls, so
  the template uses those, with the template as the import source.
- **Feed setting:** `services.chatcms.templateFeed`, defaulting to
  `https://quiqr.github.io/quiqr-community-templates/templates.json`; `null`
  switches the gallery off.
- **Freshness:** the feed is refreshed hourly. On failure the last good copy
  is served; with none, the gallery shows its error state.
- **Who sees it:** whoever may create projects in an organisation, as for the
  other two sources.
- **Repository:** the user supplies an empty repository; ChatCMS hosts none.
  The copy is not linked to the template for later updates.
- **Screenshot URL:** `templates/<NormalizedName>/screenshot.<ScreenshotImageType>`
  under the feed's own directory, as the hand-off states and the live feed
  serves.
