# Tasks

## 1. The switcher

- [x] 1.1 Add `web/src/shells/ProjectSwitcher.tsx`: a menu reading `listProjects` when it opens, listing projects that are not archived with the current one marked, a separator and "All projects"; test: `shells.test.tsx` opens it and finds both seeded projects linking to their content, the current one marked, the archived one absent, and "All projects" linking to the overview
- [x] 1.2 Say so when the list cannot be read, keeping "All projects"; test: `shells.test.tsx` with a failing fetch
- [x] 1.3 Add the strings to `en.ts` and `nl.ts`, including the switcher's name that `ProjectShell` wrote in English

## 2. The shells

- [x] 2.1 `ProjectShell`: the project tile is the switcher's trigger instead of a link
- [x] 2.2 `PhoneShell`: the project tile and name in the header are the switcher's trigger; test: `shells.test.tsx` opens it from the phone shell
- [x] 2.3 Acceptance test: from Korenaar Bakery switch to Studio Noord and land on its content, then reach the overview through "All projects"; `test/e2e/28-switching-projects.spec.mjs`

## 3. Documents

- [x] 3.1 `docs/design/DESIGN.md` section 3: what the project switcher holds
- [x] 3.2 `CHANGELOG.md` under [Unreleased]
