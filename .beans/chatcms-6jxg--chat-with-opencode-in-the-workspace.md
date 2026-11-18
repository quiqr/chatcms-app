---
# chatcms-6jxg
title: Chat with OpenCode in the workspace
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T21:41:31Z
openspec-link: openspec/changes/archive/2026-11-18-add-chat-assistant
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-18-add-chat-assistant`.

- OpenCode runs in every workspace beside the preview: a process of the
  local backend, or the guest's `chatcms-agent` unit on port 4096, which
  fetches its environment from the gateway with its token. Its whole
  configuration is passed in, a site's own is ignored, its only model is the
  gateway, its API needs the workspace's token, and it runs with no internet.
- The agent writes nothing itself: editing, commands and web access are
  denied. It changes pages through the Quiqr MCP server on the gateway
  (`list_pages`, `read_page`, `update_page`), acting as the workspace's person
  through the content service, so validation and saving are the forms'.
- A chat panel in the editor's assistant slot, and in place of the page on a
  phone: replies stream as they are written; each changed page is listed with
  Show and Undo (back to the version before, without publishing). The
  conversation survives closing the panel; the form reloads after a change.
- e2e `35-asking-the-assistant` runs the real OpenCode, gateway and tools
  against a scripted model; a real micro VM's agent was checked by hand: it
  answers with its secret, refuses without, about 10 s after start.
- OpenCode is pinned at 1.15.10 (flake input `nixpkgs-opencode`): nixpkgs now
  carries 2.0, whose server API is a different one.
- Fixed on the way: the request logging wrapper hid `Flush` and `Hijack`, so
  live reload never passed through the preview proxy (chatcms-r1wn).
- Left to the merging epic: "let the AI combine conflicting changes"
  (chatcms-izhr).
