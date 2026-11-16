# Proposal

Bean: [chatcms-x2uz](../../../.beans/chatcms-x2uz--start-without-a-repository-connect-one-later.md)
(start without a repository, connect one later).

## Why

Starting from a template asks for an empty repository before anything can
be seen, so an implementer has to go to a git host first just to try a
template. An export from Quiqr Desktop is worse off: the setup screen asks
only for its folder, so it arrives with no repository at all and can never
be published. Both should work at once and be connected to a real repository
when the implementer is ready.

## What Changes

- Starting from a template or a Quiqr Desktop export SHALL accept an empty
  git address. ChatCMS then keeps the site in a repository of its own, in its
  state directory, and editing, preview and history work straight away.
- The key step is skipped while the site is kept by ChatCMS.
- Publishing a site that is not connected yet is refused with a message that
  points to project settings, and is not recorded as a failed publish.
- Project settings can connect a site to a real repository: the key must reach
  it and it must be empty; then every version so far is sent there, the
  address is recorded, and the repository ChatCMS kept is removed.
- Removing a project also removes a repository ChatCMS kept for it.
- `docs/DEPLOYING.md` names the kept repositories under what to back up.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `onboarding/import`: a template or an export may arrive without an address
  and is then kept by ChatCMS.
- `tenancy/projects`: a kept site can be connected to a repository later, and
  removing a project removes what ChatCMS kept for it.
- `publish/act`: a site that is not connected is not published.

## Impact

- `internal/gitsync`: making a bare repository for a site and moving a site to
  a new remote. The `sync/repository` seam gains these two operations; what a
  remote is stays behind it.
- `internal/onboarding` (import, connect), `internal/tenancy` (removal),
  `internal/publish` (refusal), a migration for a `hosted` flag on projects.
- `UseTemplatePage.tsx`, `OnboardingPage.tsx`, `ProjectSettingsPage.tsx` and
  strings.

## Acceptance test served

Acceptance test 2 (import a site using a deploy key) keeps working unchanged.
This adds a second way in that needs no key until publishing.

## Assumptions

- Kept repositories live in `<state>/hosted/<project id>.git`, bare, and are
  reached as a path, so they need no key.
- Connecting accepts only an empty repository. Joining two histories is out
  of scope.
- The background sync keeps running against the kept repository, which holds
  only what ChatCMS wrote, so it finds nothing.
- Only templates and exports may start kept. A git import always has an
  address, because the address is where the site comes from.
- This reverses the hand-off's assumption 1 in
  `docs/design/template-gallery/TEMPLATE-GALLERY.md` section 6, at Pim's
  request.
