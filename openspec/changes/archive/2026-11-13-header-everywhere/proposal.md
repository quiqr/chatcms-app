# Proposal

## Why

Some screens have a header and some do not, and the person menu sits top
right on some and bottom left on others. Somebody looking for their account
or the way out has to look in a different corner depending on where they are.
Every screen should have a header, with the person menu in its top right.

Originating bean: `chatcms-3op9` (header everywhere, user menu always top
right), under milestone `chatcms-iqog` (03.1 UI cleanups).

Acceptance test served: none directly. It is a cleanup of the frame every
acceptance test in `docs/BRIEFING-M01.md` section 6 runs inside, and changes
what tests 14 (signing out), 17 (preferences) and 24 (the platform admin
area) find where.

## What Changes

- One top bar component shared by every desktop shell, with the person menu
  always as its last item.
- The project shell's sidebar loses its user block at the bottom; the person
  menu moves into its top bar, after the publish button.
- The organisation admin shell and the platform admin shell get a top bar,
  which they lacked. It is light, like every other top bar; the platform admin
  area keeps its dark sidebar as what sets it apart.
- "Back to my projects" moves from the bottom of the admin sidebars into the
  start of their top bar.
- The person menu says the viewer's role under their name, where a shell
  knows it: the role on the project, "Organisation admin" or "Platform admin".
  It replaces the name and role that stood beside the avatar in the project
  sidebar.
- The organisation shell and the editor shell use the shared top bar; their
  content does not change.
- `docs/design/DESIGN.md` section 3 is brought in line: the user is no longer
  at the bottom of any sidebar.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `interface/app-shell`: every frame has a top bar with the person menu at its
  end, and the person menu names the viewer's role.

## Impact

- `web/src/shells/`: a new `TopBar`, and `ProjectShell`, `EditorShell`,
  `OrganisationShell`, `OrganisationAdminShell`, `PlatformAdminShell` and
  `PersonMenu` changed. `PhoneShell` already has its person menu top right and
  stays as it is.
- `web/src/shells/__tests__/shells.test.tsx`: every shell, the two admin
  shells included, is checked for a person menu in its top bar.
- `docs/design/DESIGN.md` section 3 and `CHANGELOG.md`.
- `flake.nix`: the `frontend` check gives Vitest a 30 second test timeout,
  because the gate failed on timeouts in screens this change does not touch.
- No server change, no seam touched.

### Assumptions

- The phone shell is already right and is left alone.
- The admin top bars hold only the way back and the person menu; a page title
  stays in the page, as it is now.
- The editor shell's viewer carries no role, so its person menu shows the name
  only.
- Keeping "Back to my projects" visible, rather than relying on the logo,
  follows the platform admin area's requirement that its shell offers the way
  back to one's own projects.
