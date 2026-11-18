---
# chatcms-6jxg
title: Chat with OpenCode in the workspace
status: todo
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T12:51:49Z
parent: chatcms-z0p7
blocked_by:
    - chatcms-mizc
---

## Purpose

The chat panel the editor left room for: ask for a change, see it in the
preview, find it saved. Supersedes chatcms-22dh.

## Scope

- OpenCode from nixpkgs in headless server mode inside the workspace, working
  in the working copy.
- The gateway configured as OpenCode's custom provider.
- Quiqr MCP tools so the model and forms stay in sync, combined with direct
  Markdown edits.
- Backend to OpenCode over its HTTP API; replies streamed to a chat panel in
  the editor's third slot (`docs/design/DESIGN.md` sections 5.6 and 8).
- Every change listed with Show and Undo, each a commit on the user's branch.
- "Let the AI combine them" added to the conflict choice.

## Done when

"Change X on page Y" through chat changes the file, appears in the preview and
is committed on the user's branch, asserted in e2e with a stubbed model.

## References

`docs/BRIEFING-M07.md` step 3.
