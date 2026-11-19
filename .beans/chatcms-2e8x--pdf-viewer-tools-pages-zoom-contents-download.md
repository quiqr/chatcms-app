---
# chatcms-2e8x
title: 'PDF viewer tools: pages, zoom, contents, download'
status: completed
type: feature
priority: normal
created_at: 2026-11-19T16:15:46Z
updated_at: 2026-11-19T17:30:00Z
openspec-link: openspec/changes/archive/2026-11-19-pdf-viewer-tools
parent: chatcms-iqog
---

## Purpose

The PDF preview shows the pages at their own size on a grey ground, and
fits a phone's width. Reading a longer PDF needs basic viewer tools.

## Scope

- Page: "3 / 12" with previous and next, and typing a page number to go there.
- Zoom: out and in, and 100% for the real size; a phone starts fitted.
- Contents: a panel with the PDF's outline (bookmarks), which Quarto and
  Typst already write; choosing an entry goes to its page.
- Download the PDF.

## Notes

- The pane spec says a PDF preview offers no choice of width. Zoom keeps
  the format and only scales the view, so that sentence changes.
- The viewer runs inside the preview, possibly on another origin; the
  tools live in the viewer page, self-contained.

## Summary of Changes

- The PDF viewer has a sticky toolbar: the page in view out of the total,
  previous and next, and a typed page number.
- Zoom by steps from 50 to 300%, redrawn sharp, with one press back to the
  actual size; a phone starts fitted to its width.
- The PDF's outline as a contents panel; download under the file's own name.
- Labels come from the catalogue in the interface's language.
- Proven by Vitest and e2e 38 at both widths.
