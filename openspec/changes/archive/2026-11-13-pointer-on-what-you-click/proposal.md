# Proposal

## Why

Over "New project" the mouse shows the arrow, over Publish the hand. Tailwind
4's base styles give every `<button>` the arrow cursor, so whether a control
shows the hand depends on whether it was built as a link or as a button,
which nobody using the screen can know. A control that can be clicked should
always show the hand.

Originating bean: `chatcms-nerm`, under milestone `chatcms-iqog` (03.1 UI
cleanups). The bean speaks of hovering; it means the cursor, as its author
said when this was explored.

Acceptance test served: none directly. It is a cleanup of every screen the
acceptance tests in `docs/BRIEFING-M01.md` section 6 use.

## What Changes

- One base rule in `web/src/index.css` gives the pointing hand to buttons,
  menu items, tabs, options, disclosures, selects and the labels of
  checkboxes, radios and file pickers, and leaves it off when they are
  disabled.
- `MenuItem` no longer sets the arrow cursor on itself.
- The interface check fails on any component that sets `cursor-default` or
  `cursor-auto`.
- An acceptance test measures the cursor in Chromium on a button, a link, the
  person menu and a menu item, and on a disabled button.
- `docs/design/DESIGN.md` section 4 states the rule.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `interface/design-system`: what can be clicked shows the hand, and the
  frontend check enforces that nothing sets the arrow.

## Impact

- `web/src/index.css`, `web/src/components/ui/menu.tsx`,
  `web/scripts/check-interface.mjs`.
- `test/e2e/27-the-pointer.spec.mjs`, run by the `e2e` check.
- `docs/design/DESIGN.md` and `CHANGELOG.md`.
- No server change, no seam touched.

### Assumptions

- Links already show the hand in every browser and need no rule.
- A disabled button keeps whatever the browser gives it; the requirement is
  only that it is not the hand.
