---
# chatcms-ew6n
title: Convert a print design into a Quarto PDF project
status: draft
type: feature
priority: normal
created_at: 2026-11-19T09:23:13Z
updated_at: 2026-11-19T09:23:13Z
parent: chatcms-rlwc
blocked_by:
    - chatcms-j2yr
---

The same flow for print designs (HTML with print CSS): the layout becomes a Typst template, the text becomes .qmd files, and a Quiqr model covers fields such as title, author and cover image. Typst is the default, since CSS layout converts more reliably to it and it renders fast enough for a live preview. Quarto with LaTeX is for imports that already use a LaTeX template, or for academic and publisher requirements. Either is only a runtime definition per project, not a separate system.

Source: `~/secondbrain/content/quiqr_chatcms__claude_design_import_en_conversie.md`. Draft: how this is done is still to be discussed.
