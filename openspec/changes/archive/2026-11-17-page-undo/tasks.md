# Tasks

## 1. The steps

- [x] 1.1 Add `web/src/routes/undo.ts` with `record`, `planUndo`,
      `planRedo`, `planTo` and `fits`, as in design.md. Verify with
      `web/src/routes/__tests__/undo.test.ts`: one step per record, the cap of
      100, a new record emptying redo, undo and redo writing before and after,
      a changed field left alone and named, going to a point as one write,
      and `fits` dropping steps for a page changed elsewhere.
- [x] 1.2 Add sessionStorage keeping under
      `chatcms.undo.<person>.<project>.<page>`, wrapped in try/catch. Verify in
      `undo.test.ts`: steps survive a reload, another page has its own, and a
      throwing storage still leaves undo working for the open page.

## 2. The editor

- [x] 2.1 In `PageEditor.tsx`, note each path's before value when it first
      enters `waiting`, and make a step when `savePage` succeeds and none when
      it fails. Verify in `editor.test.tsx`: two quick edits give one step
      naming both fields, and a refused save gives none.
- [x] 2.2 Add the undo write path: await the save in flight, write what is
      waiting, then write the planned values without recording a step; say a
      field left alone with a toast. Verify in `editor.test.tsx`: undoing a
      cleared summary sends the old text, redo sends the new one, undo before
      the save sends the typing and then the undo, in that order.
- [x] 2.3 Add Undo, Redo and Undo history buttons beside the saved status,
      44 px, disabled when there is nothing to do, Undo and Redo named after
      what they would change; add the strings in `en.ts` and `nl.ts` with no
      git words. Verify in `editor.test.tsx`: both disabled on opening, the
      names, and `npm run check:strings` and `npm run check:interface`.
- [x] 2.4 Add the history menu: steps newest first, redoable ones dimmed above
      the current point, the current point checked, "Opened" last, times from
      `lib/when.ts`; picking a row goes there in one save. Verify in
      `editor.test.tsx`: three steps then "Opened" restores the opened values
      with one save and lists three redoable steps.
- [x] 2.5 Move `typing` from `routes/content/usePlace.tsx` to
      `web/src/lib/typing.ts`, and add the keys: Ctrl+Z or Cmd+Z undoes,
      Ctrl+Shift+Z, Cmd+Shift+Z or Ctrl+Y redoes, not in a text field. Verify
      in `editor.test.tsx`: the keys outside a field undo and redo, Ctrl+Z
      inside the body field does not undo the page, and the existing Alt+Up
      and Alt+Down tests still pass.

## 3. Documents

- [x] 3.1 Add a short undo paragraph to §5.6 of `docs/design/DESIGN.md`:
      where the buttons sit, the history list, and that an undo is a save.
      Verify that the prose hook passes on the file.
- [x] 3.2 Add a CHANGELOG entry under `[Unreleased]`, `### Added`, at most two
      lines of at most 80 characters: undo and redo in the editor, with a list
      of the steps since the page was opened. Verify by reading the section.

## 4. End to end

- [x] 4.1 Add `test/e2e/32-page-undo.spec.mjs`: clear a field, wait for it to
      be saved, undo, reload and see the old value; redo; pick "Opened" from
      the undo history; at 390 px the three buttons are there and at least
      44 px. Verify that the e2e suite passes at 1440 px and 390 px.
- [x] 4.2 Run `nix flake check` and verify it is green.

## Workflow follow-up

- Ship with `bash scripts/ship-change.sh page-undo "<subject>"`.
- Add the openspec-link to bean chatcms-6a72 and close it with a Summary of
  Changes.
