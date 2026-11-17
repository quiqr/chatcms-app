# Proposal

Bean: [chatcms-6a72](../../../.beans/chatcms-6a72--page-history-and-rollback-function.md)
(per page undo with a visual undo history).

## Why

The editor saves every change 800 ms after typing stops. Nothing is ever lost,
but nothing can be taken back quickly either: a cleared field or a removed
paragraph is saved before anybody notices, and the only way back is going back
on the whole site in History. People need an undo for the page they have
open, so that saving at once stops feeling like a risk.

## What Changes

- **Steps.** Each save the editor makes is one undo step. A step remembers
  the fields it changed, and the page text when that changed, with their
  value before and after, and when it was made.
- **Undo and redo** buttons sit next to whether the page is saved, in the
  editor at every width. Ctrl+Z or Cmd+Z undoes and Ctrl+Shift+Z, Cmd+Shift+Z
  or Ctrl+Y redoes, when the focus is not in a text field. Inside a text
  field the browser's own undo keeps working.
- **An undo is a save.** It writes the values from before the step through
  the same save as any edit, so it makes a version like any other and can
  itself be redone. Nothing is ever removed from the site's history.
- **Undo touches only its own fields.** A field that no longer holds the
  value the step left in it is not overwritten; the editor says which field
  it left alone.
- **A visual undo history.** A list opened from an "Undo history" button
  beside them shows the steps since the page was opened, newest first: the
  fields each changed, as the model names them, and the time. "Opened" is the
  last row. Picking a row brings the page back to that point in one save; the
  later steps stay as redo.
- **A new edit after an undo** drops the steps that could have been redone,
  as in any editor.
- **The steps last as long as the browser tab**, for the open page, and
  survive a reload of it. They are thrown away when the page as loaded no
  longer matches where the steps left it, and capped at 100.

## Capabilities

### New Capabilities

- `content/undo`: undoing and redoing edits to the open page, step by step or
  back to a point in its visual history.

### Modified Capabilities

None. An undo goes through the existing save, so `content/editing` keeps its
requirements: an edit changes only what it names, and every save makes a
version.

## Impact

- Frontend only: `web/src/routes/PageEditor.tsx`, a new undo module next to
  it, new strings in `en.ts` and `nl.ts`. No API, database or Go change; no
  seam in gitsync, runner, policy or auth is touched.
- Vitest in `web/src/routes/__tests__/editor.test.tsx` and a new unit test
  for the undo steps; a new Playwright spec `32-page-undo.spec.mjs` at 1440 px
  and 390 px.
- Acceptance test: it serves test 4 of `docs/BRIEFING-M01.md` section 6
  (create, edit and delete a page), and test 9 (everything works on a phone).

## Assumptions

- One step per save is the right size. The 800 ms pause already cuts typing
  into chunks the way other editors group undo.
- sessionStorage, per person, project and page, is enough. Going back further
  than this tab is page history, which the bean leaves out.
- The words for the undo controls avoid git words: "Undo", "Redo", "Undo
  history", "Opened". The list never shows a version identifier.
- A step is made only from a save that succeeded, so somebody whose saves
  are refused, such as a viewer, has nothing to undo and sees the buttons
  disabled. Hiding the editor's controls from viewers is not part of this
  change.
