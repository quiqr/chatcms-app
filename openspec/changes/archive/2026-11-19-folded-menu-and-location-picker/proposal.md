# Proposal

Bean: [chatcms-vqsf](../../../.beans/chatcms-vqsf--folded-content-menu-a-vertical-label-strip-and-a-l.md)
(folded content menu: a vertical label strip and a location picker),
milestone 03.1 ([chatcms-iqog](../../../.beans/chatcms-iqog--ui-cleanups.md)).
Design hand-off: `docs/design/folded-menu/FOLDED-MENU.md`.

## Why

Folded, the content column leaves only an unfold icon: nothing says where the
person is or what is behind it, and switching to another page means unfolding
it first. It serves acceptance test 3 of `docs/BRIEFING-M01.md` ("all forms
from the model render"): every form must stay one step away however the
screen is arranged.

## What Changes

- **A vertical label strip** replaces the folded column's bare icon. The whole
  48 px strip is one button that unfolds the column; it reads `CONTENT ·
  <current item> <count>` from top to bottom, only `CONTENT` on the content
  home, with an orange dot when anything in the menu waits to be published. A
  long title is cut short in the strip and shown whole in a tooltip.
- **A location picker** in the breadcrumb of the collection view and the
  editor, folded or not: `Content / <group> / <current ▾>`. Its last part
  opens a popover with quick find and the menu in compact form; groups of
  more than six items show only their label and count until opened; "Show
  the full content menu" unfolds the column. Keyboard: arrows move, Enter
  opens, Esc closes and gives focus back.
- The editor keeps its place in a collection ("3 of 10") beside the
  breadcrumb.

## Capabilities

### New Capabilities

### Modified Capabilities

- `content/content-menu`: the folded column is a label strip; a location
  picker reaches every page without unfolding it.

## Impact

- `web/src/routes/content/`: the strip in `ContentPanel`, a new breadcrumb
  with the picker; `usePlace` gives the position on its own.
- `web/src/components/ui/popover.tsx`: shadcn's popover over the Radix
  package already installed.
- `web/src/routes/ProjectPage.tsx`, `ProjectShell`, `EditorPage`: the content
  screens show the breadcrumb where the title was.
- Touches none of gitsync, runner, policy or auth.

## Assumptions

- **Background and colours from our tokens**: the strip uses the content
  column's own background and border, not the reference's inline hex values.
- **The picker's quick find searches as the quick find dialog does**: the
  same server search over titles, the same pause while typing.
- **On a narrow window the strip opens the content menu as a drawer**, as the
  bare icon does now; "Show the full content menu" does the same there.
- **No picker on a phone**: the phone frame keeps its own flow, as the spec
  asks.
