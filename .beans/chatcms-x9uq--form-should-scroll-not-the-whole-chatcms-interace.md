---
# chatcms-x9uq
title: form should scroll, not the whole chatcms interace
status: completed
type: task
priority: normal
created_at: 2026-11-19T11:37:44Z
updated_at: 2026-11-19T14:16:19Z
openspec-link: openspec/changes/archive/2026-11-19-panels-scroll-on-their-own
parent: chatcms-iqog
---

with long forms, the bottom of the preview window disappears without reason and the chat input as well. Only the form should scroll.

Maybe this is a more generic rule for other area's as well.

Maybe chatcms-y9bd is not needed if this is done right.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-19-panels-scroll-on-their-own`.

- Every desktop frame is as tall as the window; the page no longer scrolls.
  The header stays; the navigation, content menu, form, preview and
  assistant each scroll inside themselves, so a long form leaves the
  preview's end and the assistant's message box in view.
- The editor shows form and preview side by side only where both fit; on a
  narrower window it shows one, the form first, with the switch at every
  desktop width.
- Where form, preview and assistant do not fit, the assistant takes the
  form's place beside the preview.
- Below 1024 px the navigation shows as its rail and opens as a drawer, like
  the content menu, without changing the person's own fold.
- e2e `42-panels-scroll-on-their-own`.
