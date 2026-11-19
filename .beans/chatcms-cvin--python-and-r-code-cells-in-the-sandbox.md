---
# chatcms-cvin
title: Python and R code cells in the sandbox
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-19T07:19:37Z
parent: chatcms-q9qu
blocked_by:
    - chatcms-yu3v
---

Code cells run during preview and build inside the workspace or build sandbox,
without network or secrets. Their packages are pinned through Nix in the
runtime definition, so a render is reproducible.
