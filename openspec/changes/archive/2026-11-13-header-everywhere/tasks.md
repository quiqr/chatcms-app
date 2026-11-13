# Tasks

## 1. The top bar and the person menu

- [x] 1.1 Add `web/src/shells/TopBar.tsx`: a `<header>` with the shared top bar classes, its children, and the person menu last; test: every shell has the person menu inside its banner (`shells.test.tsx`)
- [x] 1.2 Give `PersonMenu` an optional `role`, shown under the name; test: the project shell's person menu shows the role on the project, the platform admin shell's shows "Platform admin"

## 2. The shells

- [x] 2.1 `ProjectShell`: drop the user block from the sidebar, put the top bar's person menu after the publish button with the project role
- [x] 2.2 `OrganisationAdminShell` and `PlatformAdminShell`: a light top bar beside the sidebar with "Back to my projects" and the person menu; the sidebar loses both; test: both shells in the way-out test and the top bar test, and e2e test 24 still finds "Back to my projects"
- [x] 2.3 `OrganisationShell` and `EditorShell`: use `TopBar`, content unchanged
- [x] 2.4 Check that no shell places `PersonMenu` outside `TopBar` or `PhoneShell`; test: the top bar test in `shells.test.tsx`

## 3. Documents

- [x] 3.1 `docs/design/DESIGN.md` section 3: the user sits at the end of the top bar in every shell, admin shells have a light top bar
- [x] 3.2 `CHANGELOG.md` under [Unreleased]

## 4. The gate

- [x] 4.1 The `frontend` check runs Vitest with a 30 second test timeout: under the load of `nix flake check` ten untouched screen tests ran past the default 5 seconds; local runs keep the default
