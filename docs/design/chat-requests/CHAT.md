# Phase 1 — Chat from the preview

Editors and implementers talk to the assistant **from the preview**: select
text, say what should change, follow progress at the selection and undo it
there. A fourth view, **Chat + preview**, shows the whole conversation next to
the page.

References: `reference/phase-1-chat/screenshots/` and the same states as static
HTML (exact inline styles) in `reference/phase-1-chat/html/`. Use the existing
tokens (`tokens/tokens.css`) and shadcn components, not the inline hex values.
Sample content (ecosymbia-like case study) is illustrative only.

Scope of phase 1: §1–§7. Not in this phase: requests between people (phase 2,
`REQUESTS.md`), pointing at images (later), point mode for blocks (later).

| # | Reference | State |
|---|---|---|
| 01 | `01-selection-chip` | Chip after selecting text |
| 02 | `02-selection-input` | Input below the selection |
| 03 | `03-selection-input-above` | Input flipped above (no room below) |
| 04–07 | `04-progress-*` … `07-progress-several` | Working · done · question · several at once |
| 08 | `08-chat-and-preview` | Chat + preview view with a request card |
| 09–11 | `09-template-element-*`, `11-shared-content` | Selection inside template / shared content |
| phone | `phone-01` … `phone-03` | Phone (secondary priority) |

---

## 1. Who gets it

- **Editors and implementers**: yes.
- **Viewers**: no chip, no chat in phase 1. (Phase 2 gives them "Request a
  change".)
- Only in **Preview** and **Chat + preview**. Not in Form or Form + preview.

## 2. The overlay script

The preview is the real Hugo site on its own origin. A small script injected
into preview builds draws everything on top of the site.

- **Self-contained.** One script, its own styles in a shadow root, no
  dependency on the site's CSS or JS. It must look like Quiqr, not like the
  site: Geist, the app tokens, a 1 px `rgba(24,24,27,.14)` border and the
  menu shadow, so nobody mistakes it for site content.
- **Talks to the app** via `postMessage` with a fixed origin check in both
  directions. The app owns state (chat stream, undo); the overlay only shows
  it.
- Messages (names are suggestions):
  - overlay → app: `selection:send {text, prompt, anchor}`, `undo {changeId}`,
    `open-chat {threadId}`
  - app → overlay: `change:state {changeId, state, summary}`,
    `highlight {anchor}`, `badge {count}`
- **Anchor** of a selection = page URL, build id, the closest
  `data-quiqr-field` (see §6), the selected text, and a text offset inside that
  field. The assistant gets the anchor with the prompt.

## 3. Chip after selecting (01)

- Show a small dark pill **"Q Chat"** at the end of the selection
  (`aria-label="Chat about the selected text"`), 6 px below the last line.
- Appears **after** the selection ends, never while dragging:
  mouse — on `mouseup` + 150 ms; keyboard — 400 ms after the last
  shift-arrow; touch — selection handles still for 500 ms.
- No chip for selections shorter than 2 characters or across more than one
  field (`data-quiqr-field`).
- Disappears when the selection is cleared, on Esc, or on scroll out of view.
- `Ctrl/⌘ + Shift + K` with a selection opens the input directly.

## 4. Input at the selection (02, 03)

Clicking the chip turns it into the input card (380 px):

- Textarea **"What should change?"**, round send button (36 px, `primary`).
- Quick actions as chips: **Make it shorter · Simpler words · Fix spelling**.
  A chip sends immediately with that prompt.
- Footer, 11 px `muted-foreground`: "Chat · goes to your chat" and
  "Enter to send · Esc to close". Shift+Enter = new line.
- **Placement**: 12 px below the selection with an arrow pointing at it,
  clamped 12 px from the preview edges. If it doesn't fit below, it **flips
  above** (03). Follows the selection while scrolling; if the selection
  scrolls out, it collapses to the chip at the edge of the preview.
- **Esc or click outside** closes it. Typed text is kept as a draft in the
  chat box (Chat + preview), not thrown away.
- The selection stays highlighted (`primary` 20 % tint) while the card is
  open.

## 5. Progress and Undo at the selection (04–07)

After sending, the input becomes a pill anchored to the same place:

- **Working** — "Working on it…" + **Stop**.
- **Done** — "Changed · **Undo** · Show in chat". The changed text is briefly
  tinted `success`. The pill stays at least 8 s and until the person
  interacts with the preview again; after that Undo lives in the chat only.
- **Question** — "The assistant has a question · **Answer in chat**". Clicking
  opens Chat + preview with the question focused.
- **Several at once** (07) — each selection keeps its own pill; the
  **Chat + preview** segment in the view switch shows a badge with the number
  of running or unanswered items.
- Undo restores the field to its exact previous value and adds a line to the
  chat ("Undone"). No git words anywhere for editors: "changed", "undo",
  "saved".

## 6. Content or template? (09–11)

Selections are mapped to content via `data-quiqr-field` attributes that the
preview build adds around rendered fields (title, body, front-matter fields,
shared/global settings). Three cases:

1. **Inside a page field** → normal chip and input (§3–4).
2. **Inside shared content** (global settings, menus, data files) (11) → the
   card says "**Shared text** · Global settings › Footer — appears on all
   24 pages", keeps the input, and has **Open Global settings › Footer**.
   Changes apply to every page; say so before sending.
3. **No field: part of the template** (09, 10):
   - **Editor** (09): lock icon, "**Part of the site design.** Same on every
     page." In phase 1, show this message only — the request field and
     "Send request to implementers" arrive in phase 2.
   - **Implementer** (10): code icon + the template file and line
     (e.g. `layouts/partials/footer.html:18`) and an input that goes to their
     own chat. The "open requests for this element" link is phase 2.
- How to know the template file: in preview builds only, a partial adds a
  `data-quiqr-tpl="<file>:<line>"` hint (check whether `templates.Current`
  in our Hugo version is enough). Fallback: the assistant searches
  `layouts/` and the theme for the element's classes and text.

## 7. Chat + preview view (08)

- The view switch becomes **Form · Form + preview · Chat + preview ·
  Preview**. Remember the last choice per person per project.
- Left: a **400 px chat column**, preview on the right.
- The chat stream is **private per person per project**. No one else sees it.
- A message sent from the preview shows as a **request card**: a snippet of
  the page with the selection highlighted, the prompt, and the caption
  "From the preview · <page title>". Clicking the card scrolls the preview to
  the anchor and highlights it.
- Each change the assistant makes gets a **change summary** with **Show
  changes** and **Undo**. Show changes highlights it in the preview (green
  tint) with a "Changed · Undo" pill.
- Questions from the assistant appear as normal messages; the composer at the
  bottom ("Reply or ask a change").
- There is no separate "Assistant" button anywhere: the Chat + preview
  segment carries the badge.

## 8. Phone (< 640 px, secondary)

Tabs **Edit · Preview · Chat** instead of the view switch. Selecting text in
Preview opens a **bottom sheet** with the quoted selection and the input;
done/question show as a pill above the tab bar; the chat tab shows the same
request cards. Touch targets ≥ 44 px. Do this after desktop works.

## 9. Acceptance checks

1. Selecting text in Preview shows the Chat chip only after the selection
   ends; no chip while dragging or for a single character.
2. The input opens below the selection, flips above near the bottom edge, and
   stays inside the preview with 12 px margin.
3. Esc closes the input and the typed text is in the chat composer.
4. Sending shows Working → Done with Undo at the selection; Undo restores the
   exact previous text.
5. A question shows "Answer in chat" and opens Chat + preview at that
   question.
6. Two running changes give two pills and a badge "2" on Chat + preview.
7. Chat + preview shows the request card; clicking it scrolls the preview to
   the selection.
8. Selecting inside the footer shows the template card (editor: lock text;
   implementer: file and line). Selecting in a global setting shows the
   shared-content card.
9. Viewers see no chip.
10. The overlay renders identically on two different Hugo themes and doesn't
    pick up the site's fonts or styles.
11. No git words in any editor-facing copy.
