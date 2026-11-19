# Phase 2 — Requests

People ask each other for changes: **viewers → editors → implementers**.
Editors are the first line; they forward to implementers when something needs
a design change. Requests are shared with the people they go to — unlike the
private chat from phase 1.

Requires phase 1 (overlay, anchors, chat + preview, template detection).

References: `reference/phase-2-requests/screenshots/` and static HTML in
`reference/phase-2-requests/html/`. Tokens and shadcn as in phase 1.

| # | Reference | Who · state |
|---|---|---|
| 01–02 | `01-viewer-chip`, `02-viewer-request` | Viewer · request a change from the preview |
| 03–04 | `03-editor-for-me-*` | Editor · a request for me (open, first version) |
| 05–06 | `05-editor-sent-*` | Editor · a request I sent (picked up, done) |
| 07–08 | `07-editor-chat-card-*` | Editor · request card in my chat |
| 09–10 | `09-implementer-for-me-*` | Implementer · inbox (open, picked up) |
| 11–12 | `11-code-preview-*` | Implementer · Code + preview (normal, theme warning) |

---

## 1. Data model

A **request** has:

- `from` (a person), `to` (a **role**: `editors` or `implementers`;
  optionally also a person),
- `about`: an anchor from phase 1 (page, element, selected text, build id)
  **or** a page **or** nothing,
- `type`: `content` · `template` · `question` (extensible; later
  `check-before-publish`),
- a **thread** of messages, and an **activity log** (sent, picked up,
  forwarded, done, declined, reopened),
- `status`: `open` · `picked_up` · `done` · `declined`,
- `owner`: whoever picked it up (no separate assignment),
- a **snapshot** (§6).

Forwarding moves the **same** request to another role; thread and history
stay. It never creates a copy.

Visibility: the sender, everyone in the `to` role, and everyone it passed
through. Show this on every input: "Visible to Joost and editors".

## 2. Requests box (all roles)

- New rail item **Requests** with a badge (open items for me).
- Tabs **For me** and **Sent by me**; filters Open/Done and type.
- List item: thumbnail of the snapshot (or an icon when `about` is empty),
  title, type label, sender, origin when forwarded ("Joost → via Sanne"),
  status, age.
- Detail: header with type · "to <role>", title, status pill; the snapshot
  with "as it looked when sent" and a link to the page; the thread; actions;
  the reply field with its visibility line.
- "Ticket" never appears in the UI. Words: Request, Send request to…,
  Picked up, Forwarded, Done, Declined.

## 3. Viewer (01–02)

- In Preview, selecting text shows **"Q Request a change"** instead of Chat
  (same timing rules as phase 1 §3).
- Card: textarea "What should change?", button **Send request to editors**,
  footer "Visible to you and the editors · reply comes in My requests". No
  quick actions, no assistant.
- Viewers always send **to editors** and never choose a type or recipient.
- Header for viewers: a "Viewer" label and **My requests** (the Requests box,
  Sent by me only). No view switch, no Publish.

## 4. Editor

**For me (03–04)** — actions on an open request:

- **Forward to assistant** — the content assistant makes a first version
  (e.g. a new page in the suggested collection, or the text change). The
  draft is only visible to this editor until publish. The request card
  goes into the editor's own Chat + preview.
- **I'll write it myself** — picks it up and opens the page.
- **Forward to implementers** — with an optional note. The sender sees
  "Forwarded by Sanne".
- **Decline…** — with a required short reason, sent to the sender.
- The assistant suggests a **place** for content requests ("Suggested place:
  Pages · new page").
- A content request is marked done when the related change is published.
  **Ask <sender> to check it** sends the preview link to the sender first.

**Sending (template element, phase 1 §6 case 3)** — the lock card gets a text
field and **Send request to implementers** with "Visible to implementers".

**Sent by me (05–06)** — tracker *Sent → Picked up by <name> → Done*, the
snapshot, the thread. When done: **Show in preview** and **Not quite
right…** (reopens the same request with a message).

**In the chat (07–08)** — a sent request also appears in the sender's own
chat as a card with a lock icon, "Request to implementers · Footer · Open",
and a status card that updates. The preview marks the element "Request open"
(amber dashed outline) and later "Changed by <name>".

## 5. Implementer

**For me (09–10)** — template requests and forwarded requests.
Actions: **Forward to template assistant**, **I'll do it myself**,
**Decline…**. Both of the first two pick the request up (owner = me; the
sender sees my name) and open **Code + preview**.

**Code + preview (11–12)** — a fifth view, **implementers only**:
Form · Form + preview · Chat + preview · **Code + preview** · Preview.

- Left: chat with the **template assistant** (may change `layouts/`,
  `assets/`, config, archetypes, the Quiqr model). The request card sits at
  the top.
- Middle: the changed files as a diff, tabs per file; footer with the
  preview build status.
- Right: preview of the **template draft**, labelled "Template draft · only
  you see this".
- Template work happens on its own draft (branch), separate from editors'
  content edits. Git words are fine here.
- Change summary in the chat: files, +/−, "Affects N pages · goes live with
  the next publish", **Discard** / **Apply**. Apply merges into what the next
  publish takes and marks the request done; the sender gets a message.
- **Warnings** only when they mean something, shown in the chat above the
  summary:
  - file belongs to a **theme or Hugo module** → "Changes inside
    themes/<name> are lost on the next theme update." Default action **Copy
    to layouts/ and change it there**; secondary "Change the theme file
    anyway". Apply disabled until one is chosen (12).
  - the change **removes or renames a field** the Quiqr model or a
    `data-quiqr-field` uses → requires "Apply anyway".
  - the **preview build fails** or shows new errors → Apply blocked.
  - the change touches a **base template or a partial used everywhere**
    while the request was about one element → say so, offer a narrower one.
  - another implementer has an **open draft on the same file** → say who.

## 6. Snapshot of the element

Captured when a request is sent from the preview:

1. **In the browser (overlay)**: anchor (phase 1 §2) + CSS selector path,
   bounding box, viewport size, trimmed `outerHTML`, visible text,
   `data-quiqr-*` hints. Cheap, no permissions.
2. **On the server, async**: headless Chromium opens the same preview
   **build id** at the same viewport, clips the bounding box (+16 px) and
   stores a PNG. Real fonts and images, frozen in time.
- Fallback if no server browser: store the extracted HTML with inlined
  computed styles and render it in a sandboxed iframe (no scripts,
  sanitised).
- Don't use html2canvas-style client rendering.

## 7. Email

Viewers and many editors won't have the app open, so requests need email.

**Needs**
- Outgoing transactional mail from a project-neutral sender (e.g.
  `requests@chatcms.eu`), SPF/DKIM/DMARC set up. Use the platform's existing
  mail provider if there is one.
- **Reply by email**: `Reply-To: r+<request-token>@chatcms.eu`. Inbound
  mail is parsed (strip quoted text and signatures), checked against the
  sender's address, and added to the thread. Unknown sender or bad token →
  bounce politely, add nothing.
- Per person, per project setting: **Every update · Daily summary · Off**
  (default: every update for viewers, daily summary for editors and
  implementers). One-click unsubscribe link in every mail.
- Never put the private chat in an email. Only the request thread, the
  snapshot PNG and links.

**When**
| Event | Who gets mail |
|---|---|
| New request to my role | everyone in that role (respecting their setting) |
| Forwarded to my role | everyone in that role |
| Picked up · Forwarded · Done · Declined | the sender (and people it passed through) |
| New reply in the thread | everyone on the request except the writer |

**Copy (example, new request to editors)**
- Subject: `Joost asked for a change on "Protecting the Lear's macaw"`
- Body: the snapshot image, the request text, "Reply to this email to
  answer Joost.", button **Open request**, footer "You get this because
  you're an editor on ecosymbia.com · Change email settings".

## 8. Acceptance checks

1. A viewer selects text and sends a request; it appears in every editor's
   For me with the snapshot, and the viewer sees it in My requests.
2. An editor forwards it to implementers with a note; the implementer sees
   "Joost → via Sanne"; the thread is the same; Joost sees "Forwarded by
   Sanne".
3. "Forward to assistant" on a content request creates a draft visible only to
   that editor; publishing marks the request done and informs the sender.
4. An editor sends a template request from the lock card; it appears in the
   implementers' inbox, and as a card + status in the editor's chat.
5. An implementer forwards to the template assistant; Code + preview shows the
   diff and the draft preview; editors' previews are unchanged until Apply +
   publish.
6. Changing a file under `themes/` shows the theme warning and Apply stays
   disabled until a choice is made.
7. A failing preview build blocks Apply.
8. Snapshots stay the same after the page changes.
9. Every input shows who can see it.
10. Replying to a notification email adds the reply to the thread; an email
    from another address is rejected.
11. Email settings work per project; unsubscribe works with one click.
12. No git words in viewer or editor copy.
