# Tasks

## 1. The fold state

- [x] 1.1 Add `web/src/lib/useFold.ts`: `[folded, setFolded]` for a panel name
      and the person, over `recall` and `remember`, with an optional fallback
      key read when the own key is absent. Verify with a Vitest in
      `web/src/lib/__tests__/useFold.test.tsx`: the default is unfolded, a
      change survives a remount, and the fallback key is honoured.

## 2. The site panel

- [x] 2.1 Add `web/src/shells/SitePanel.tsx` with the full form (logo,
      project switcher, Site and Setup groups, fold icon in the heading) and
      the folded rail (unfold icon on top, project tile opening the switcher,
      the same entries as icon links with an accessible name and a tooltip,
      the publish count kept). Its outer element carries `data-panel="site"`
      and `data-folded`. Add the strings for folding and unfolding the
      navigation in `en.ts` and `nl.ts`. Verify with Vitest in
      `shells.test.tsx`: both forms, setup entries only for owner and
      implementer, and `npm run check:strings`.
- [x] 2.2 Render `SitePanel` in `ProjectShell` and in `EditorShell`, removing
      the editor's own rail and `railEntries`. Verify with Vitest in
      `shells.test.tsx`: folding on a project screen and then rendering the
      editor shows the rail; unfolded on a project screen shows the full
      navigation in the editor; the fold survives a remount.

## 3. The content panel

- [x] 3.1 Add `web/src/routes/content/ContentPanel.tsx` around
      `ContentColumn`: the 256 px column with its fold icon, or the 48 px
      strip with one icon that unfolds at 1024 px and wider and opens the
      `ContentDrawer` below that. It carries `data-panel="content"` and
      `data-folded`, and reads `chatcms.content.folded.<person>` as fallback.
      Verify with Vitest in `contentmenu.test.tsx`: fold and unfold at
      1440 px, the strip opening the drawer at 800 px, and the old key giving
      a folded start.
- [x] 3.2 Render `ContentPanel` in `ContentScreen` and `EditorShell`, removing
      the collection view's outline drawer button and the rail's Content
      button; the rail's Content entry is a link like the others. Remove
      strings that are no longer used. Verify with Vitest: folding on the
      collection view then opening the editor shows the strip, and
      `npm run typecheck`.

## 4. The motion

- [x] 4.1 Add the `panel-motion` block to the end of `web/src/index.css`
      between its two marker comments, as in design.md, and nothing else for
      motion anywhere. Verify by hand at 1440 px: folding moves, a reload does
      not, and with reduced motion set in the browser nothing moves; and by
      `grep -rn "transition" web/src --include=*.tsx` showing no panel
      transition in a component.

## 5. Documents

- [x] 5.1 In `docs/design/DESIGN.md`, replace the editor exception with "the
      editor shows the panels as the person left them", add a short motion
      section holding the rule and the way to undo it, and add the override
      line to §5.5. Verify that the prose hook passes on the file.
- [x] 5.2 Add the CHANGELOG entries under `[Unreleased]`, each at most two
      lines of at most 80 characters: the navigation and the content menu
      fold by hand and stay as left (Changed), and opening a page no longer
      folds the navigation (Fixed). Verify by reading the section.

## 6. End to end

- [x] 6.1 Update `test/e2e/31-content-menu.spec.mjs`: step 4 folds and
      unfolds through the strip; a new step folds the navigation on the
      collection view, opens an item, sees the rail, reloads and still sees
      it, and unfolds it. Update any spec that looked for `editor-rail`.
      Verify that the e2e suite passes at 1440 px and at 390 px.
- [x] 6.2 Run `nix flake check` and verify it is green.

## Workflow follow-up

- Ship with `bash scripts/ship-change.sh panels-fold-by-hand "<subject>"`.
- Add the openspec-link to beans chatcms-jmy8 and chatcms-tpop and close both
  with a Summary of Changes.
