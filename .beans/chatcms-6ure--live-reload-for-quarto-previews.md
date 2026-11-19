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
parent: chatcms-z0p7
---

Quarto's preview script finds its websocket only under its own proxy prefixes (`/p/<word>/`, `/proxy/<port>/`). Under `/v/<key>/` and `/preview/<org>/<project>/` it falls back to the root and never connects, so a Quarto preview only refreshes when the pane reloads it after a save. Changes made outside a save, and re-renders that take longer than the reload, are not seen.

Found in `openspec/changes/archive/2026-11-19-add-quarto-runtime`.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-19-fix-quarto-previews`. Quarto's websocket already reached the preview; after a
render it went to `/index.html` at the host's root, which answered 404.
Rooted previews are now addressed as `<prefix>p/site/`, which Quarto keeps,
and the proxy takes both off. Proven by e2e `38-a-quarto-book`: a chapter
changed on disk shows in the pane without a save.
