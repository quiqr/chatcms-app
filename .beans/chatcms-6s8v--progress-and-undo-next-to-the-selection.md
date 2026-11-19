---
# chatcms-6s8v
title: Progress and Undo next to the selection
status: completed
type: feature
priority: normal
created_at: 2026-11-19T09:20:47Z
updated_at: 2026-11-20T00:30:00Z
openspec-link: openspec/changes/archive/2026-11-19-progress-at-the-selection
parent: chatcms-z0p7
blocked_by:
    - chatcms-7hfz
---

A request made on the canvas shows its status next to the selection, "working on it" then "done", with an Undo button there. When the assistant asks something back, a badge near the selection points to the chat.

Source: `~/secondbrain/content/quiqr_chatcms__chat_interfaces.md`.

## Summary of Changes

A request sent from the selection shows a pill there: working with Stop,
changed with Undo and Show in chat (the new text tinted green), or a question
with Answer in chat. Done pills stay at least 8 seconds and until the next
touch; pills of several requests stack; they come back after a reload of the
preview. Undo from the pill puts back all of that request's changes in one
step; Stop aborts the assistant on the server or drops a waiting request. The
Chat + preview count includes unanswered questions.
