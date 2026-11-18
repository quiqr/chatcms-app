# Tasks

## 1. The script in the preview

- [x] 1.1 `preview.Decorate`: inject into HTML, serve `/__chatcms/`, leave
  the rest; used by the app's preview route and the preview origin; proven
  by preview tests.
- [x] 1.2 `select.js`: the button on a selection, the message to the app,
  the highlight on request; proven by a Vitest over the script in jsdom.

## 2. The app side

- [x] 2.1 The pane takes selections from its frame only and passes them up;
  sends the highlight after a change; proven by Vitest.
- [x] 2.2 The chat panel holds a quote and sends it with the message; the
  editor opens the assistant on a selection, on a phone too; proven by
  Vitest.

## 3. Finding the page

- [x] 3.1 `model.KeyOf`, shared with `sitesync`; proven by model tests.
- [x] 3.2 `/chat` takes a selection; `Chat.Send` builds the prompt; the
  `Locator` with `hugo list all`; proven by assistant, httpapi and app tests.

## 4. PDF

- [x] 4.1 The PDF.js viewer build, embedded and served under `/__chatcms/pdf/`;
  the pane frames it; the page kept on refresh; selections posted; proven by
  Vitest and the Nix build.

## 5. Acceptance

- [x] 5.1 e2e: select a sentence on a Hugo page, ask, the right page changes
  and the new text is highlighted; a pdf preview shows its text through
  PDF.js; at both widths. A message from elsewhere is ignored: Vitest.

## 6. Ship

- [x] 6.1 `docs/ARCHITECTURE.md`, `CHANGELOG.md`.
- [x] 6.2 `scripts/gate.sh ship` green.
