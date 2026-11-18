# Design

## Context

See proposal.md. Previews pass through `internal/preview`: the app's own
`/preview/{org}/{project}/` behind the session, or `/v/<view>/` on the
preview origin. Both forward the whole path to the workspace, whose Hugo
serves under that prefix. The assistant edits pages through the MCP tools
on the gateway; the chat panel posts `{text}` to `/chat`.

## Decisions

### 1. One decorated proxy for both addresses

`preview.Decorate(proxy, prefix, app)` wraps the reverse proxy for a
preview reached under `prefix`:

- `prefix + "__chatcms/select.js"` and `prefix + "__chatcms/pdf/…"` are
  served from files embedded in the binary, never from the workspace;
- the request's `Accept-Encoding` is dropped so HTML comes back plain;
- an answer with `Content-Type: text/html` and status 200 gets
  `<script src="{prefix}__chatcms/select.js" data-app="{app origin}" defer></script>`
  before `</body>`, case-insensitively, or at the end; `Content-Length` is
  set again. Everything else passes as it came.

### 2. The messages

From the preview, to `data-app` only:

```json
{ "chatcms": "selection", "text": "…", "before": "…", "after": "…", "path": "/v/<view>/posts/one/" }
```

From the app, to the preview's origin only:

```json
{ "chatcms": "highlight", "before": "…", "after": "…" }
```

The pane checks `event.source === frame.contentWindow` and
`event.origin === new URL(frame.src).origin`. The script checks
`event.origin === data-app`. The highlight uses the CSS Custom Highlight
API over the body's text, between the end of `before` and the start of
`after`, for four seconds; where the API is missing nothing is shown.

### 3. Address to page

`Locator.Locate(ctx, key, path)` in `internal/app`: the person's copy, the
project's resolved runtime; if its preview program is `hugo`, run
`hugo list all --source <copy>` with the runtime's `PATH`, read the CSV
(`path`, `permalink`, …), compare each permalink's path, with the site's
base path removed, to the asked path with the preview prefix removed; the
matching file goes through `model.KeyOf`. Anything else gives no page.

### 4. The prompt

`Chat.Send` with a selection adds, after the person's words:

```
The person selected this on the preview page /posts/one/:
"<selected text>"
Just before it: "<before>"
Just after it: "<after>"
It is on the page with key "posts/one". Read it with read_page, change it
with update_page as asked, and change nothing else.
```

or, without a page: `No page was found for this address. Find the page that
holds this text with list_pages and read_page; if no page holds it, say it
comes from the site's design, which you cannot change.`

### 5. The PDF viewer

A second Vite build (`web/vite.pdf.config.ts`, entry `web/pdf/index.html`,
`base: "./"`) writes `dist/pdf/`. It reads `?file=` (a path under the same
prefix) and `#page=`, renders pages with their text layer, remembers the
page in `sessionStorage` per file, and posts selections like the web
script. The pane frames `<preview>/__chatcms/pdf/?file=<open>`.
