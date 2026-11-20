---
# chatcms-1jm8
title: An editor undo test times out under a full Vitest run
status: completed
type: bug
priority: low
created_at: 2026-11-19T07:46:28Z
updated_at: 2026-11-20T21:13:27Z
openspec-link: openspec/changes/archive/2026-11-20-steadier-under-load
---

`web/src/routes/__tests__/editor.test.tsx`, "undo > goes back to the opened page in one save from the undo history", passed alone every time but timed out after 5 s once during a full `npx vitest run` on a loaded machine. The gate runs Vitest with a 30 s timeout, so it has not failed there yet.

## Summary of Changes

- The page editor's quiet time before saving can be shortened; the undo
  tests use 20 ms instead of waiting 800 ms four times.
