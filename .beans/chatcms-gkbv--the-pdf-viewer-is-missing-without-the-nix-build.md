---
# chatcms-gkbv
title: The PDF viewer is missing without the Nix build
status: completed
type: bug
priority: low
created_at: 2026-11-19T07:46:28Z
updated_at: 2026-11-19T10:49:20Z
---

A plain `go build` embeds a placeholder frontend, so `/__chatcms/pdf/` and `/__chatcms/select.js` answer 404 and a pdf preview shows nothing. `scripts/dev-be.sh` builds this way, so PDF previews and select-to-edit do not work in development unless the frontend is built first. Either the dev server serves them from Vite, or dev-be builds the preview assets.

Found in `openspec/changes/archive/2026-11-19-add-select-to-edit`.

## Summary of Changes

`scripts/dev-be.sh` builds the preview assets (`vite.preview.config.ts`) and
puts them in `internal/webui/dist/__chatcms`, which is ignored by git, before
it compiles; so select-to-edit and PDF previews work in development. Without
`web/node_modules` it says they are off and why.
