updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
updated_at: 2026-11-19T08:07:42Z
openspec-link: openspec/changes/archive/2026-11-19-fix-quarto-previews
---

The micro VM overlays the site, but the local backend runs the preview in the person's copy itself. Anything Quarto writes that the site's `.gitignore` does not cover makes the copy dirty, and a dirty copy stops bringing in other people's changes, without saying why. The Quarto book fixture ignores `/.quarto/` and `/_book/`; a site that forgets is stuck.

Found in `openspec/changes/archive/2026-11-19-add-quarto-runtime`.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-19-fix-quarto-previews`. Status and the fast-forward of a person's copy ignore
files git does not track, so a preview's output is never taken for unsaved
work and no longer stops other people's changes. When an arriving change adds
a file a preview wrote, the preview's file gives way.
