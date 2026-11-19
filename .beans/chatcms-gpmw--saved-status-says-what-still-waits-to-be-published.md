---
# chatcms-gpmw
title: Saved status says what still waits to be published
status: todo
type: feature
priority: normal
created_at: 2026-11-19T22:49:44Z
updated_at: 2026-11-19T22:49:44Z
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
