---
# chatcms-6ure
title: Live reload for Quarto previews
status: todo
type: task
priority: normal
created_at: 2026-11-19T07:46:28Z
updated_at: 2026-11-19T07:46:28Z
parent: chatcms-z0p7
---

Quarto's preview script finds its websocket only under its own proxy prefixes (`/p/<word>/`, `/proxy/<port>/`). Under `/v/<key>/` and `/preview/<org>/<project>/` it falls back to the root and never connects, so a Quarto preview only refreshes when the pane reloads it after a save. Changes made outside a save, and re-renders that take longer than the reload, are not seen.

Found in `openspec/changes/archive/2026-11-19-add-quarto-runtime`.
