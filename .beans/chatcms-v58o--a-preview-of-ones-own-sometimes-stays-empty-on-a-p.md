---
# chatcms-v58o
title: A preview of one's own sometimes stays empty on a phone
status: completed
type: bug
priority: low
created_at: 2026-11-19T07:19:37Z
updated_at: 2026-11-20T21:13:27Z
openspec-link: openspec/changes/archive/2026-11-20-steadier-under-load
---

`test/e2e/33-a-preview-of-ones-own.spec.mjs`, "two people each see their own change in the preview", failed once on the phone pass of the release gate (full suite, under load): the first person's preview frame kept an empty body for 90 s, while the second person's showed the site. It passed 16 of 16 when run alone at both widths, and in the ship gate right after.

Seen while shipping `openspec/changes/archive/*-add-quarto-runtime`.

## Summary of Changes

- The overlay reports an empty page with `ready`; the preview loads it again
  at most three times. The PDF viewer never counts as empty. e2e 33 passed
  every run at 12 workers.
