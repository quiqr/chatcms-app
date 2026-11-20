---
# chatcms-slum
title: A folder of documents, one PDF rendered on request
status: todo
type: feature
priority: normal
created_at: 2026-11-20T15:03:46Z
updated_at: 2026-11-20T15:03:46Z
---

A Quarto project may hold many documents that each become a PDF, such as
a folder of quotations. Rendering them all on a preview start costs far too
much, so one document at a time is rendered when it is asked for.

Today the preview runs one command for the whole site and `preview.open` is
a single fixed path, so the pane shows one PDF and starting it renders
everything.

The shape agreed on 20 November 2026:

- Two placeholders: `{file}`, the source document of the page being edited,
  and `{page}`, the same path with `.pdf` after it. `preview.open` may use
  `{page}` so the pane follows the page.
- A `render` command, run on request rather than at start:
  `[quarto, render, "{file}"]`. Without it nothing changes for existing
  sites.
- An endpoint that runs it in the workspace (in the copy for the local
  backend, inside the sandbox for gVisor), streams the log to the project
  log, and refreshes the pane when it succeeds.
- One render at a time per project, with cancelling and a time limit; a
  second request waits or says the first is still running.
- The result stays in the workspace, so returning to a quotation costs
  nothing. When the source is newer than the PDF the pane says it is out of
  date and offers to render again.
- Publishing is unchanged: a publish still builds everything.

Chosen on 20 November 2026: a button, not automatic. Rendering on simply
opening a page would spend the same effort on every visit that the automatic
preview does now.
