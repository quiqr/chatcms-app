---
# chatcms-ljgu
title: A content menu test times out under load
status: completed
type: bug
priority: low
created_at: 2026-11-18T20:29:59Z
updated_at: 2026-11-20T11:45:17Z
---

`web/src/routes/__tests__/contentmenu.test.tsx`, "an open item's place >
says where it is and steps through the list as it was", ran past Vitest's
5 s timeout once in the `frontend` check while the release gate built
everything else alongside it. On its own it passes every time.

Expected: the test does not depend on the machine being idle, either by
waiting for what it needs rather than a fixed time, or by a longer timeout for
this test with the reason written down.

## Summary of Changes

- The test now presses the key inside its wait, so it no longer depends on
  the listener being attached in time (commit 819953a).
