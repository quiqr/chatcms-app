# Tasks

## 1. The viewer's frame

- [x] 1.1 Labels from the catalogue and `&lang=` on the viewer's address from
  the pane; proven by Vitest.
- [x] 1.2 The sticky toolbar, 44 px targets, styled in `viewer.css`.

## 2. The tools

- [x] 2.1 Pages: "n / total" following the page in view, previous, next, and
  a typed number clamped to the PDF; proven by Vitest on the helpers.
- [x] 2.2 Zoom: steps out and in from any start, 100%, redraw from the page
  in view outwards, the page kept in view, an older redraw abandoned; proven
  by Vitest on the steps.
- [x] 2.3 Contents from the outline, nested, beside the pages or over them
  on a phone; hidden without an outline.
- [x] 2.4 Download as a link to the file.

## 3. Proof and ship

- [x] 3.1 e2e 38 "previewed as a PDF": page field, next, zoom in and back to
  100% with text still selectable, contents to a chapter, download link, at
  both widths.
- [x] 3.2 `CHANGELOG.md` under [Unreleased].
- [x] 3.3 `scripts/gate.sh ship` green; close chatcms-2e8x.
