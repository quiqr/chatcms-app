# Proposal

## Why

The project at the top of the project sidebar carries an up-and-down chevron,
so it looks like a menu to switch projects, but it is a link to the overview
of all projects. Somebody moving between two sites has to go through the
overview every time, and on a phone the project in the header does nothing at
all.

Originating bean: `chatcms-dx31` (switch project button should really switch
project), under milestone `chatcms-iqog` (03.1 UI cleanups).

Acceptance test served: none directly. It shortens the way between projects
that acceptance tests 1 and 11 in `docs/BRIEFING-M01.md` section 6 walk
through the overview.

## What Changes

- The project at the top of the desktop sidebar opens a menu: the projects in
  this organisation, the current one marked, then a separator and "All
  projects".
- The project in the phone header opens the same menu.
- Choosing a project opens its content; "All projects" opens the overview, as
  the link did before.
- Archived projects are not offered.
- The switcher's accessible name comes from the string catalogue; it was
  written in English in the component.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `interface/app-shell`: the project switcher switches projects, on the
  desktop and on the phone.

## Impact

- `web/src/shells/ProjectSwitcher.tsx` (new), `ProjectShell.tsx`,
  `PhoneShell.tsx`.
- `web/src/strings/en.ts` and `nl.ts`: the switcher's name, "All projects",
  and what it says while loading or when the list cannot be read.
- `web/src/shells/__tests__/shells.test.tsx` and a new acceptance test.
- `docs/design/DESIGN.md` section 3 and `CHANGELOG.md`.
- No server change: the menu reads the same project list the overview does.
  No seam touched.

### Assumptions

- The menu lists projects of the current organisation only; moving between
  organisations stays with the organisation switcher on the overview.
- The list is read each time the menu opens, so a project made or archived
  meanwhile is right without a reload.
- The editor keeps its icon rail and has no switcher.
