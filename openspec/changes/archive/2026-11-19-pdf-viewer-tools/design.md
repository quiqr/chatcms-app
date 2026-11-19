# Design

## Context

The PDF viewer is its own small page, `web/src/preview/pdf/`, built by
`vite.preview.config.ts` into `__chatcms/pdf/` and served under the preview's
address, possibly on another origin than the app. It draws every page with
PDF.js to a canvas with a text layer, remembers the page in view per file in
session storage, and installs the select-to-edit script when framed by the
app. `scaleFor` decides the scale: own size, or fitted when the address says
`fit=width`.

## Goals / Non-Goals

**Goals:** page position and navigation, zoom by steps, the outline as
contents, download; labels in the interface's language; selection keeps
working.

**Non-Goals:** search in the PDF, thumbnails, rotation, printing, remembering
the zoom.

## Decisions

### 1. The toolbar lives in the viewer page

The pane cannot reach into a frame on another origin, so the tools are part
of the viewer page itself: a sticky bar above `#pages`, in plain DOM like the
rest of the viewer, styled in `viewer.css` with the design tokens' values
copied as custom properties. The bar is 44 px high, every control a 44 px
target, so it passes the phone checks.

### 2. Zoom redraws, it does not CSS-scale

A zoom step sets a new scale and redraws every page's canvas and text layer
at it, the way the first draw already does. CSS transforms would blur the
canvas and shift the text layer against it. The page in view is kept by
remembering its number before the redraw and scrolling to it after. The
steps are 0.5, 0.75, 1, 1.25, 1.5, 2 and 3 times the own size; a fitted
phone start sits between them, and the next step in either direction is the
nearest one past it. Redrawing is done page by page from the one in view
outwards, so the page being read is sharp first.

### 3. Pages tracked by the existing observer

The IntersectionObserver that remembers the page in view also updates the
page field. Previous, next and a typed number call `scrollIntoView` on the
page's box; the number is clamped to 1..n.

### 4. Contents from `getOutline`

`pdf.getOutline()` gives the bookmarks; each destination is resolved with
`getDestination` and `getPageIndex` when chosen. The panel is a nested list
of buttons beside the pages on a wide pane (260 px) and over them on a phone,
closing after a choice. An empty outline hides the button.

### 5. Download is a link

An `<a download>` to the file's own address: same origin as the viewer, so
the browser saves it under its name without fetching it again in script.

### 6. Labels from the catalogue, language from the address

The viewer imports the catalogue (`strings/en.ts`, `strings/nl.ts`) and picks
the language from `&lang=` on its address, which `Preview.tsx` adds from the
language in force. The select-to-edit label keeps arriving by message as it
does now. Importing the catalogue makes the viewer bundle larger by the
catalogue's size (tens of kilobytes, cached); the alternative, sending every
label by message, would leave the viewer unlabelled until the app answers and
when it is opened on its own.

## Risks / Trade-offs

- **Redrawing a long book at 300%** takes time and memory. Pages are drawn
  from the one in view outwards and a redraw in progress is abandoned when
  another zoom comes, so the page in view is always the first to be sharp.
- **Highlight and select at other zooms**: the text layer is rebuilt with the
  same viewport as the canvas, which is how the first draw keeps them
  aligned; checked in e2e at 100% and 150%.
