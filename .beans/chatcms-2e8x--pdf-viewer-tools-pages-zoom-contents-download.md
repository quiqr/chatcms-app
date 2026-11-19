---
# chatcms-2e8x
title: 'PDF viewer tools: pages, zoom, contents, download'
status: in-progress
type: feature
priority: normal
created_at: 2026-11-19T16:15:46Z
updated_at: 2026-11-19T16:15:46Z
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
