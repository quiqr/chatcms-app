---
# chatcms-r1wn
title: Live reload does not pass through the preview proxy
status: todo
type: bug
priority: normal
created_at: 2026-11-18T14:50:03Z
updated_at: 2026-11-18T14:50:03Z
---

## What happens

The server log repeats `http: proxy error: can't switch protocols using
non-Hijacker ResponseWriter type *httpapi.statusRecorder`. The request
logging middleware wraps the response writer without passing on
`http.Hijacker`, so the reverse proxy cannot upgrade Hugo's live-reload
websocket, and `/preview/.../livereload` answers 502.

## Effect

A preview does not reload by itself when the site changes. It shows the
change on Refresh or after a save, which is what the e2e tests do, so
nothing catches it.

## Expected

`statusRecorder` passes `Hijack` (and `Flush`) through to the writer it
wraps, live reload connects through the app, and a test proves an upgrade
reaches the preview.
