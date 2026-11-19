---
# chatcms-2myi
title: The local backend lets a preview write into the person's copy
status: in-progress
type: bug
priority: normal
created_at: 2026-11-19T07:46:28Z
updated_at: 2026-11-19T07:46:28Z
---

The micro VM overlays the site, but the local backend runs the preview in the person's copy itself. Anything Quarto writes that the site's `.gitignore` does not cover makes the copy dirty, and a dirty copy stops bringing in other people's changes, without saying why. The Quarto book fixture ignores `/.quarto/` and `/_book/`; a site that forgets is stuck.

Found in `openspec/changes/archive/2026-11-19-add-quarto-runtime`.
