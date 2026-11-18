---
# chatcms-64h1
title: Runtime definition and one preview pane
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T22:25:27Z
openspec-link: openspec/changes/archive/2026-11-18-add-runtime-definition
parent: chatcms-z0p7
blocked_by:
    - chatcms-6jxg
---

## Purpose

A site declares what it needs to run, and one preview pane shows whatever that
produces. The ground for Quarto in the next epic.

## Scope

- A declarative runtime definition in the site's quiqr directory: packages and
  versions (nixpkgs only, simple mode), preview and build commands, preview
  type (web, pdf, other). The Hugo pin from the micro VM epic moves into it.
- The preview and build commands are the site developer's; ChatCMS provides
  runtime, sandbox, lifecycle, proxy and output handling.
- Previews served from an origin separate from the app, so site JavaScript can
  never reach the ChatCMS app.
- One preview pane: web in an iframe with live reload, pdf in the browser's
  own viewer, other as a download.

## Done when

- A Hugo site with a runtime definition previews as before, from the separate
  origin.
- A site whose preview type is pdf shows a PDF in the pane.

## Open decisions

- Exact format of the runtime definition file. Flag in the proposal.

## References

`docs/BRIEFING-M07.md` step 4 and the addendum "Generic preview pane".

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-18-add-runtime-definition`.

- `quiqr/runtime.yaml` declares packages with versions, a preview command
  with its type (web, pdf, other) and page to open, and a build with its
  output. Packages come through nixpin; without the file a site runs Hugo as
  before. Commands are word lists with placeholders the backend fills in;
  a wrong declaration refuses the preview with a reason.
- Both backends run the declared command; the micro VM gets it, filled in,
  as `chatcms.run` on its command line. Checked by hand on a real VM: cold
  start about 6.4 s. Builds run the declared command and publish its output.
- `CHATCMS_PREVIEW_URL` (`services.chatcms.previewUrl`) moves previews to an
  origin of their own, reached at `/v/<key>/` with a random key per
  workspace and the person's right checked on every request; the app's own
  `/preview/` stops serving.
- The pane shows a web page or a PDF, and offers anything else to open.
- Found by the e2e: a running preview ignored a changed definition; opening
  the preview now starts afresh when the file's hash changed.
- e2e `36-a-preview-apart` covers the separate origin and a PDF preview at
  both widths. Reference in `docs/RUNTIME.md`.
