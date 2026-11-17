# Proposal

Beans: [chatcms-jmy8](../../../.beans/chatcms-jmy8--auto-toggling-site-panel-is-confusing.md)
(the site panel toggling by itself is confusing) and
[chatcms-tpop](../../../.beans/chatcms-tpop--the-sidepanels-need-collapse-icons-and-uncollapse.md)
(both side panels need icons to fold and unfold them). Both sit under the
milestone chatcms-iqog, 03.1 UI cleanups.

## Why

Opening an item turns the project's navigation into a 64 px rail by itself,
and leaving the item turns it back. Nobody asked for it, so it reads as the
screen jumping. At the same time the person has no way to fold the navigation,
and the content column folds only inside the editor. The panels change when
the screen wants and never when the person wants: that is backwards.

## What Changes

- **Only the person folds a panel.** No screen changes a panel's state.
  Opening an item, going back to a collection or moving to Media keeps both
  panels as they were.
- **The site panel folds to an icon rail.** Its heading gets a fold icon. The
  folded panel is the 64 px rail with the same entries as icons, labelled for
  screen readers and with a tooltip, and an unfold icon at the top. This is
  the same in the project screens and in the editor.
- **The content panel folds to a thin strip.** Its heading gets a fold icon,
  on the collection view as well as in the editor. The folded panel is a 48 px
  strip holding one unfold icon, in the place where the panel was.
- **Both choices are remembered per person**, across projects and reloads.
  Somebody who folded the content column in the editor before this change
  keeps it folded.
- **A short animation** shows where a panel goes when the person folds or
  unfolds it: about 180 ms, no motion with reduced motion set, and none when a
  screen first loads. All of it lives in one place in the stylesheet, so
  taking it out is a deletion of a few lines and nothing else changes.
- **BREAKING (design):** the hand-off's rule that the editor collapses the
  sidebar to a rail (`docs/design/content-menu/CONTENT-MENU.md`, "Item open")
  and the editor exception in `docs/design/DESIGN.md` are dropped. The editor
  shows the panels as the person left them.
- Below 1024 px the content panel stays a drawer, and the phone frame does
  not change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `interface/app-shell`: the editor no longer swaps the navigation for a rail
  by itself; the navigation folds and unfolds by hand in every project frame,
  with a short animation that respects reduced motion.
- `content/content-menu`: the content column folds on the collection view as
  well as in the editor, into a strip with its own unfold icon, instead of
  coming back only from the rail.

## Impact

- Frontend only. No API, database or Go change, and no seam in gitsync,
  runner, policy or auth is touched.
- `web/src/shells/ProjectShell.tsx` and `EditorShell.tsx` share one site
  panel with a full and a folded form. `ContentScreen.tsx` and
  `EditorShell.tsx` share one content panel with a full and a folded form.
- One new block of motion rules in the stylesheet; `DESIGN.md` gains its
  first rule on motion.
- Strings for folding and unfolding the navigation, in `en.ts` and `nl.ts`.
- Vitest: `shells.test.tsx` and `contentmenu.test.tsx`. Playwright: spec 31,
  step 4, and the specs that look for the editor's rail.
- Acceptance test: none of section 6 of `docs/BRIEFING-M01.md` changes. The
  change serves test 4 (create, edit and delete a page) by making the editor
  stop moving under the person's hands, and keeps test 9 (a phone) as it is.

## Assumptions

- The fold state is per person, not per project: a person who wants room
  wants it everywhere. The content column's groups stay per project, as now.
- The editor's first visit shows both panels unfolded, like every other
  screen. At 1440 px this leaves about 940 px for form and preview side by
  side, which works; one fold frees 400 px more and is remembered.
- The old key `chatcms.content.folded.<person>` is read once as the content
  panel's starting state, so nobody's choice is lost.
- Between 640 and 1023 px the site panel folds the same way; under about
  800 px it still moves above the content as the reflow rule says.
