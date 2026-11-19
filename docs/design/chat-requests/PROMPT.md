# Prompts for Claude Code

Put this folder in the repo as `docs/design/chat-requests/`.

## Phase 1 — Chat from the preview

```
New feature for Quiqr ChatCMS: chat with the assistant from the preview.
Read docs/design/chat-requests/CHAT.md first. References are in
docs/design/chat-requests/reference/phase-1-chat/ (screenshots + static HTML
with exact inline styles). Use our tokens and shadcn components, not the
inline hex values.

Build §1–§7 in this order:
1. The overlay script for preview builds (§2): shadow root, postMessage
   protocol with origin checks, anchors via data-quiqr-field.
2. Chip and input at the selection (§3, §4).
3. Progress, Undo and questions at the selection (§5).
4. The Chat + preview view with request cards and change summaries (§7).
5. Content / shared / template detection and the three cards (§6). For
   editors, the template card only shows the message in this phase.

Leave the phone layout (§8) for last and tell me before you start it.
Don't build anything from REQUESTS.md yet. Make a short plan first and show
it to me; when done, show screenshots next to the references and run the
acceptance checks in §9.
```

## Phase 2 — Requests

```
Phase 2 of chat in Quiqr ChatCMS: requests between viewers, editors and
implementers. Read docs/design/chat-requests/REQUESTS.md (phase 1 is
CHAT.md and is already built). References are in
docs/design/chat-requests/reference/phase-2-requests/.

Build in this order:
1. Data model, permissions and visibility (§1), with tests.
2. The Requests box for all roles (§2) and the viewer flow (§3).
3. Editor actions, forwarding and the chat card (§4).
4. Snapshots (§6): overlay capture first, then the server-side PNG.
5. Implementer inbox and Code + preview with the template assistant,
   template drafts and the warnings (§5).
6. Email: notifications, settings, reply by email (§7). Ask me which mail
   provider and inbound setup to use before you start this step.

Make a short plan first and show it to me. When done, run the acceptance
checks in §8.
```
