---
# chatcms-64h1
title: Runtime definition and one preview pane
status: in-progress
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T21:47:41Z
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
