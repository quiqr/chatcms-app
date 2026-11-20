---
# chatcms-gpmw
title: Saved status says what still waits to be published
status: completed
type: feature
priority: normal
created_at: 2026-11-19T22:49:44Z
updated_at: 2026-11-20T12:53:47Z
openspec-link: openspec/changes/archive/2026-11-20-saved-beside-the-form
parent: chatcms-iqog
---

## Purpose

The editor's sync indicator says "All changes saved" while the Publish button
counts changes not yet published and the content menu marks changed pages.
Both are true, but side by side they read as a contradiction to an editor:
saved is not live.

## Idea

Let the indicator say both when something waits: "Saved · 2 changes not
published yet", and plain "Saved" when nothing waits. No git words.

## Done when

- The indicator names the number waiting to be published, and agrees with the
  Publish button, at 1440 px and 390 px.

## Summary of Changes

- Solved differently than the idea above, agreed on 20 November 2026: the
  top bar's save status, which always said "All changes saved", is gone.
- Whether a page is saved is said only beside the form: "Saved",
  "Saving…" or not saved yet. The Publish button keeps the count.
