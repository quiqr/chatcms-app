---
# chatcms-vqsf
title: 'Folded content menu: a vertical label strip and a location picker'
status: in-progress
type: feature
priority: normal
created_at: 2026-11-18T15:04:32Z
updated_at: 2026-11-19T14:28:51Z
parent: chatcms-iqog
---

Folded, the content column leaves only an unfold icon: nothing says what is
behind it or where you are. The design hand-off in `docs/design/folded-menu/`
adds two things that work together. The spec is
`docs/design/folded-menu/FOLDED-MENU.md`; screenshots and static HTML of both
states are in its `reference/`.

- **A vertical label strip** of 48 px replaces the folded column. The whole
  strip is one button that unfolds the column. It reads `CONTENT · <current
  item> <count>` from top to bottom, only `CONTENT` on the content home, with
  an orange dot when anything in the menu has unpublished changes. A long
  title is truncated, shown in full in a tooltip. The fold state stays
  remembered per person.
- **A location picker** in the breadcrumb of the collection view and the
  editor, folded or not. Its last part opens a 320 px popover with the same
  quick find as the column and the model menu in compact form; groups of more
  than six items show only their label and count until opened. "Show the full
  content menu" unfolds the column. Fully usable by keyboard; Esc closes and
  returns focus. No picker below 640 px; the phone flow stays.

Nothing else in the content area changes. Use the project tokens and shadcn
components, not the inline hex values of the references, and every string
from the catalogue.

## Done when

The eight acceptance checks in section 3 of `FOLDED-MENU.md` hold, proven by
Vitest tests and an e2e test at 1440 px, with screenshots of both states next
to the references.

## References

- `docs/design/folded-menu/FOLDED-MENU.md`, `PROMPT.md`
- Builds on the content column from chatcms-1hg1 and the fold icons of
  chatcms-tpop.
