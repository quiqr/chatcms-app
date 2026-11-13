# Design

## Context

There are six shells in `web/src/shells/`. Three put the person menu at the
end of a top bar (organisation, editor, phone). The project shell has a top
bar but keeps the person menu at the bottom of its sidebar, next to the name
and role. The two admin shells have no top bar at all and keep the person
menu, under "Back to my projects", at the bottom of their sidebar.
`docs/design/DESIGN.md` section 3 describes exactly that, so the design
document changes with the code.

## Goals / Non-Goals

**Goals:**
- One place for the person menu on every desktop screen: the end of the top
  bar.
- One component for that top bar, so a future shell cannot put it elsewhere.

**Non-Goals:**
- Changing what the sidebars navigate to, or the phone shell.
- Page titles in the admin top bars.

## Decisions

**A `TopBar` component owns the person menu.** It renders the `<header>`
with the shared classes, then its children, then the person menu with
`ml-auto`. A shell passes what goes before it and the viewer. The person menu
cannot then be placed anywhere but last, and it stays right-aligned when the
bar wraps. The alternative, a convention each shell follows, is how six
shells came to differ.

**The role moves into the person menu.** The project sidebar showed name and
role beside the avatar; a top bar has no room for that. `PersonMenu` takes an
optional `role` and shows it under the name at the top of the dropdown. The
project shell passes the project role, the admin shells pass "Organisation
admin" and "Platform admin", the organisation shell and editor shell pass
none.

**"Back to my projects" moves to the start of the admin top bars.** The
platform admin requirement asks the shell to offer the way back, and test 24
looks for the link by name. Putting it at the start of the top bar keeps both,
and leaves the sidebar with only navigation and, for the organisation admin,
the usage box.

**The admin top bars are light.** `bg-card` with a bottom border, as
elsewhere. The dark sidebar alone marks the platform admin area.

## Risks / Trade-offs

- [The role is one click further away in the project shell] → It is
  reference, not something acted on; the menu shows it at the top.
- [Tests that found the person menu inside the sidebar break] → The shell
  tests check the top bar for every shell instead; the e2e tests find the
  menu by test id, which does not change.
