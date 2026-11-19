---
# chatcms-2x4v
title: Requests between viewers, editors and implementers
status: todo
type: epic
priority: normal
created_at: 2026-11-19T19:29:25Z
updated_at: 2026-11-19T19:29:25Z
parent: chatcms-z0p7
blocked_by:
    - chatcms-xxhd
    - chatcms-7hfz
    - chatcms-6s8v
---

## Purpose

Phase 2 of the chat design: people ask each other for changes, viewers to
editors to implementers. Editors are the first line and forward to
implementers when something needs a design change. Requests are shared with
the people they go to, unlike the private chat of phase 1.

## Scope

As in `docs/design/chat-requests/REQUESTS.md`:

- The request data model, permissions and visibility.
- A Requests box for all roles, and the viewer's "Request a change".
- Editor actions, forwarding, and the request card in the chat.
- Snapshots of the element a request is about.
- The implementer's inbox and Code + preview with a template assistant,
  template drafts and warnings.
- Email: notifications, settings and reply by email (ask which mail provider
  and inbound setup first).

## References

- Design: `docs/design/chat-requests/` (REQUESTS.md, PROMPT.md phase 2,
  `reference/phase-2-requests/`).
- Requires phase 1: overlay, anchors, Chat + preview, template detection.
