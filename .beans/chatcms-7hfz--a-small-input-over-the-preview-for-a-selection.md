---
# chatcms-7hfz
title: A small input over the preview for a selection
status: completed
type: feature
priority: normal
created_at: 2026-11-19T09:20:46Z
updated_at: 2026-11-19T23:30:00Z
openspec-link: openspec/changes/archive/2026-11-19-chat-at-the-selection
parent: chatcms-z0p7
---

Selecting in the preview opens a small input above the canvas, at the selection, instead of only quoting the selection into the chat panel (what chatcms-vy58 shipped). When the person accepts, the request is added to their chat stream. It works in every view with a preview, whether or not the chat column is showing.

Source: `~/secondbrain/content/quiqr_chatcms__chat_interfaces.md`.

## Summary of Changes

Selecting text in the preview shows a Chat chip once the selection is
finished; the chip opens an input at the selection with quick actions, which
flips above when there is no room. Sending adds the request to the chat and
asks the assistant without changing the view; Esc keeps the typed text as a
draft with the selection quoted. Editors and implementers only, in Preview and
Chat + preview, and only where an assistant is available.
