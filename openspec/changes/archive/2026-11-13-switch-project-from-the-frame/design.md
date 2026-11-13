# Design

## Context

`ProjectShell` renders the project tile as a `NavLink` to `/projects` with a
`ChevronsUpDown` icon. `PhoneShell` renders the tile and name as plain text.
`listProjects(organisation)` in `lib/tenancy.ts` already returns the projects
the viewer can reach, with an `archived` flag. The `Menu` components wrap
Radix's dropdown menu and are used by the person menu and the organisation
switcher.

## Goals / Non-Goals

**Goals:**
- One switcher component used by both shells.
- No new server endpoint.

**Non-Goals:**
- Switching organisations from inside a project.
- A search field in the menu; an organisation has few projects.

## Decisions

**`ProjectSwitcher` takes its trigger as children.** The sidebar and the
phone header draw the project differently; the switcher owns the menu and
its contents and renders whatever it is given as the trigger, through Radix's
`asChild`. The trigger becomes a `button`, so it also gets the pointer and
keyboard behaviour of one.

**The list is read when the menu opens.** Reading it with every screen would
add a request to every page inside a project for a menu that is rarely
opened. While it is being read the menu says so; if it cannot be read it says
that, and "All projects" still works, because it needs nothing from the list.

**Items are links.** Each project and "All projects" is a `NavLink` inside a
`MenuItem`, as the organisation switcher does, so a project can be opened in
a new tab and the address is visible.

## Risks / Trade-offs

- [The first open shows "Loading" for a moment] → The list is small and the
  same request the overview makes; a cached list would show stale projects
  after one is archived.
