# Design

## Context

Two shells draw the project's navigation in two different ways:
`ProjectShell` has the full 240 px sidebar with no fold, and `EditorShell` has
its own 64 px rail. Moving between them is what makes the navigation change
by itself. The content column is drawn twice as well: `ContentScreen` puts it
in the page without a fold, and `EditorShell` puts it in an aside that folds
and comes back from a rail button. The fold is stored under
`chatcms.content.folded.<person>` through `lib/remember.ts`.

## Goals / Non-Goals

**Goals:**

- One site panel and one content panel, each drawn by one component that
  both shells use, so the two screens cannot disagree.
- The motion can be removed by deleting one block of CSS, with no change to
  components, tests or specs other than the motion requirement.

**Non-Goals:**

- The phone frame, the drawer's own look, and the folding of groups inside
  the content column.
- Dragging a panel to a width of the person's choice.

## Decisions

### One component per panel

`shells/SitePanel.tsx` draws the navigation in two forms: full (logo, project
switcher, the Site and Setup groups) and folded (a rail with the same entries
from `navigation.ts` as icon links with a tooltip, the project tile opening
the switcher, and the unfold icon at the top). `ProjectShell` and
`EditorShell` both render it; the editor's own `railEntries` go.

`routes/content/ContentPanel.tsx` wraps `ContentColumn`. Unfolded, it is the
256 px column with the fold icon in its heading (`onHide`, which exists
already). Folded, and always below 1024 px, it is a 48 px strip with one icon:
at 1024 px and wider the icon unfolds the column, below that it opens the
existing `ContentDrawer`. `ContentScreen` and `EditorShell` both render it, so
the "Show the content menu" outline button on the collection view and the
rail's Content button both go. The rail's Content entry becomes a link to the
content screen, like every other entry.

Alternative: keep two shells and pass a `folded` prop through. Rejected,
because the bug is two drawings of the same thing, and a prop keeps both.

### The state

A small hook `lib/useFold.ts` returns `[folded, setFolded]` for a name and the
person, over `recall` and `remember`:

| Panel   | Key                                 | Fallback key read once              |
| ------- | ----------------------------------- | ----------------------------------- |
| site    | `chatcms.panels.site.<person>`      | none                                |
| content | `chatcms.panels.content.<person>`   | `chatcms.content.folded.<person>`   |

Each shell mounts the panel afresh on a route change and reads the key on
mount, so no shared store is needed. The state is per person, not per
project; see the proposal's Assumptions.

### The motion, kept in one place

Each panel's outer element carries `data-panel="site"` or
`data-panel="content"` and `data-folded`, which the tests use anyway. Its
width comes from a Tailwind class that changes with the state, and its
content switches at once. The only motion is this block at the end of
`web/src/index.css`, between two marker lines:

```css
/* panel-motion: start */
@layer base {
  [data-panel] {
    transition:
      flex-basis 180ms ease-out,
      width 180ms ease-out;
    overflow: hidden;
  }
  @media (prefers-reduced-motion: reduce) {
    [data-panel] {
      transition: none;
    }
  }
}
/* panel-motion: end */
```

The two marker comments are required by this design, so that the block can
be found and removed whole. Undoing the decision is: delete the block,
remove the motion requirement from `interface/app-shell` and the motion line
from `DESIGN.md`. No component code, test or other requirement changes.

Why this holds:

- No JavaScript waits on the motion: no timers, no `transitionend`, no
  animation library. The state changes when the button is pressed, so the
  panel is usable at once and the tests do not see the motion at all.
- A CSS transition does not run on first paint, so a load or reload shows the
  remembered state without movement and needs no "skip on mount" code.
- `overflow: hidden` keeps the switched content from spilling while the width
  catches up. Without the block it does nothing, because the width is final
  at once.
- The block sits in `@layer base`, so a utility such as the content column's
  `overflow-y-auto` still wins over it.

Alternatives: fading labels before the width moves needs both forms mounted
at once or a delayed switch, which puts motion into the components and makes
undoing it a code change. A motion library is a dependency for 180 ms.
Both rejected.

### Design documents

`docs/design/DESIGN.md` loses the editor exception (the editor shows the
panels as they were left) and gains a short section on motion holding the
rule above. §5.5 gets one line saying that the panel rule here overrides the
hand-off's "Item open" collapse. The hand-off in `docs/design/content-menu/`
stays as delivered, because it is reference material.

## Risks / Trade-offs

- [The editor is tight at 1024 to 1300 px with both panels unfolded] → the
  form and preview already wrap below each other through the slots' flex
  basis, and the fold icon is in plain sight. No automatic fold, by design.
- [Width transition on a panel holding many links can stutter on slow
  devices] → it runs only below the person's own press, for 180 ms, and
  phones never see it.
- [Specs that look for the editor's rail by `data-testid="editor-rail"`] →
  they move to `data-panel="site"`; the e2e specs are updated in the same
  group as the shells.

## Migration Plan

Nothing to deploy beyond the frontend. The old content key is read once as
the content panel's starting state; it is left in place, so going back to an
earlier release keeps working.
