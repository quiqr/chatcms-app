# Design

## Context

See proposal.md and `docs/design/folded-menu/FOLDED-MENU.md`. `ContentPanel`
already renders a 48 px aside (`basis-12`) holding the unfold icon when
folded, or always below 1024 px, where unfolding opens `ContentDrawer`.
`ContentColumn` loads the menu (`readContentMenu`) and marks the current
entry with `isCurrent`. `QuickFind` is a dialog searching with
`searchContent`. Content screens put the word "Content" in the top bar;
the editor shows `place.context` above its title.

## Decisions

### 1. The strip

The folded aside becomes one `<button>` filling the strip, with the unfold
icon in a 36 px muted square, then the label in `writing-mode: vertical-rl`,
then the dot. The current entry comes from the same menu and `isCurrent`
the column uses, exported for the strip. The title is `truncate` with
`max-height: calc(100% - 120px)`; the tooltip carries the whole title. The
label `CONTENT` is the catalogue's word in capitals by CSS, so Dutch reads
`INHOUD`.

### 2. One way to unfold from elsewhere

"Show the full content menu" lives in the breadcrumb, far from
`ContentPanel`'s state. It dispatches a window event the panel listens for,
as `openQuickFind` already does: wide, the panel unfolds (and remembers it);
narrow, it opens the drawer.

### 3. The breadcrumb and the picker

`ContentBreadcrumb` reads the menu and the location: the group of the current
entry, and the entry itself. Its last part is a Radix popover trigger. The
popover holds a search field (focused on open; typing searches after 200 ms,
the dialog's pause) and, while the field is empty, the compact menu. Arrow
keys move a roving focus over the visible entries; Down from the field moves
into the list. Radix gives Esc and the return of focus.

ProjectShell takes an optional `heading` in place of its `<h1>`; the content
routes pass the breadcrumb. The editor passes it as its context line, with
the position (`3 of 10`) after it; `usePlace` gives that on its own.

## Risks / Trade-offs

- **The menu is fetched by the column, the strip and the breadcrumb**; three
  requests of a small JSON per screen. Sharing it would need a cache this app
  does not have; left as is.
- **Vertical text and ellipsis**: Chromium and Firefox cut vertical text with
  an ellipsis along the block; checked in the e2e at 1440 px.
