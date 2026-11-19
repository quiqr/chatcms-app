# Proposal

Bean: [chatcms-2e8x](../../../.beans/chatcms-2e8x--pdf-viewer-tools-pages-zoom-contents-download.md)
(PDF viewer tools: pages, zoom, contents, download), milestone 03.1
([chatcms-iqog](../../../.beans/chatcms-iqog--ui-cleanups.md)).

## Why

A PDF preview shows its pages at their own size, on a grey ground, and fits a
phone's width. A book of any length is then hard to read: there is no way to
see which page is showing, to go to a chapter, to look closer at a detail or
to take the PDF away. People who edit a Quarto book check its PDF here, so the
viewer needs the tools every PDF reader has.

## What Changes

- **A toolbar** above the pages of the PDF viewer, staying in place while the
  pages scroll.
- **Pages:** "3 / 12" follows the page in view; previous and next buttons, and
  typing a number goes to that page.
- **Zoom:** out and in by steps, and a button back to 100%, the page's own
  size. A desktop starts at 100%; a phone starts fitted to its width. Zoom
  scales the view only; the page keeps its format and proportions.
- **Contents:** a panel with the PDF's outline (its bookmarks), nested as in
  the PDF. Choosing an entry goes to its page. Without an outline the button
  is not there.
- **Download:** the PDF as it is.
- The viewer says its labels in the interface's language.

## Capabilities

### New Capabilities

- `preview/pdf-viewer`: the tools of the PDF viewer.

### Modified Capabilities

- `preview/pane`: a PDF preview starts at its own size, or fitted on a phone,
  and offers zoom instead of offering no choice of size.

## Impact

- `web/src/preview/pdf/`: the toolbar, the contents panel and zoom in the
  viewer page; its stylesheet.
- `web/src/routes/Preview.tsx`: the viewer's address carries the language.
- `web/src/strings/`: the viewer's labels.
- Text selection and the remembered page keep working at every zoom.
- Touches no Go code.

## Assumptions

- Zoom steps are 50, 75, 100, 125, 150, 200 and 300 percent of the page's own
  size, plus the fitted size on a phone. Zoom is not remembered between
  visits; the page in view is, as today.
- Pinch zoom on a phone is left to the browser.
- No search inside the PDF and no thumbnails in this change.
