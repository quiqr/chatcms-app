# Quiqr ChatCMS — chat from the preview & requests

Two phases:

1. **Chat** — talk to the assistant from the preview: select text, ask a
   change, follow progress and undo at the selection, and a Chat + preview
   view. Spec: `CHAT.md`.
2. **Requests** — viewers → editors → implementers, with a Requests box,
   forwarding, a template assistant in Code + preview, and email
   notifications with reply by email. Spec: `REQUESTS.md`.

How to use:

1. Copy this folder into the repo as `docs/design/chat-requests/`.
2. Start Claude Code and paste the phase 1 prompt from `PROMPT.md`. Use the
   phase 2 prompt when phase 1 is done.

| Path | What it is |
| --- | --- |
| `CHAT.md` | Phase 1 spec, numbered sections, acceptance checks. |
| `REQUESTS.md` | Phase 2 spec, incl. snapshots and email. |
| `PROMPT.md` | Kick-off prompts for both phases. |
| `tokens/tokens.css` | Design tokens (same as the main kit). |
| `reference/phase-1-chat/` | 11 desktop states + 3 phone states, PNG and static HTML. |
| `reference/phase-2-requests/` | 12 desktop states, PNG and static HTML. |
| `reference/canvas-source/` | The Claude Design artboards (`*.dc.html`). |

Not included (later): pointing at images, point mode for whole blocks,
phone layouts for requests. Sample content (ecosymbia-like) is illustrative
only.
