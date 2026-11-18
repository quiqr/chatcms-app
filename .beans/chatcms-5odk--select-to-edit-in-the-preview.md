---
# chatcms-5odk
title: Select-to-edit in the preview
status: todo
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T21:47:03Z
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
