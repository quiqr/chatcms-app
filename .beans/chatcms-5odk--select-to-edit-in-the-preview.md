---
# chatcms-5odk
title: Select-to-edit in the preview
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T23:23:39Z
openspec-link: openspec/changes/archive/2026-11-19-add-select-to-edit
parent: chatcms-z0p7
blocked_by:
    - chatcms-64h1
---

## Purpose

Select text in the preview, say what should change, and the agent changes the
source.

## Scope

- The proxy injects a small script into HTML previews: on selection, a
  floating icon; selection and surrounding context go to the parent by
  origin-checked postMessage.
- The icon opens an inline prompt or quotes the selection into the chat panel.
- Preview URL mapped to content file with `hugo list all`; fallback, the agent
  searches the repo.
- The agent edits the source, the preview reloads, the changed text is
  highlighted briefly; a short-lived undo or view change.
- PDF (Quarto): selection through PDF.js, the agent searches chapter sources.

## Done when

Selecting a sentence on a Hugo page and asking for a change edits the right
content file, shows the change highlighted and leaves a commit on the user's
branch.

## References

`docs/BRIEFING-M07.md`, addendum "Select-to-edit".

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-19-add-select-to-edit`.

- Every HTML page a preview passes on gets `<preview>/__chatcms/select.js`,
  built from `web/src/preview/` and embedded. On a selection it shows a
  button; pressing it posts the text, its context and the address to the
  app's origin only. The pane takes it only from its own frame.
- The selection opens the assistant with the text quoted (desktop and
  phone); the message carries it. For a Hugo site, `hugo list all` in the
  build sandbox maps the address to a content file and `model.KeyOf` to
  the page key; the assistant is told the page, or to look for the text.
- After the change the reloaded preview highlights the new text for four
  seconds. Show and Undo stay on the change.
- PDF previews are shown by a PDF.js viewer under the preview's own address:
  selectable text, the page kept on refresh, selections sent the same way.
- Found on the way: the chat panel could send before its event stream was
  open and miss the reply; it now holds the message until the stream is
  ready.
- e2e `37-select-to-edit` and the PDF part of `36-a-preview-apart`, at
  both widths. No highlight in PDFs; left to the Quarto epic.
