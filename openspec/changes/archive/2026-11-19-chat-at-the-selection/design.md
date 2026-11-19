# Design

## Context

`select.ts` is the overlay injected into preview pages (and the PDF viewer):
a shadow-rooted round button that posts `selection` to the app, and a
`highlight` it draws with the CSS Custom Highlight API. `Preview.tsx` sends
it `setup` with one label and forwards `selection` to `EditorPage`, which
quotes it into `ChatPanel` and switches to Chat + preview.

## Decisions

### 1. The overlay owns the chip and the card; the app owns the chat

The overlay draws the chip and the card in its shadow root, in plain DOM
with the app's tokens copied as custom properties (Geist from the app's
origin is not reachable from a site on another origin, so the system font
stack stands in). It posts three messages: `send {text, prompt}`,
`draft {text, prompt}` and, for the PDF viewer and older callers,
`selection`. The app answers `setup {words, on}`: the catalogue's words in
the interface's language, and whether chat is on (role and view). `on` is
sent again whenever the view changes.

### 2. When a selection is finished

`selectionchange` alone cannot tell dragging from done. The overlay tracks
the last input: a pointer held down shows nothing; `pointerup` waits 150 ms;
a key with Shift waits 400 ms after the last one; touch waits 500 ms after
the last `selectionchange`. Each new event restarts the wait.

### 3. Placement

The card is 380 px wide (less on a narrow preview), placed 12 px under the
selection's last line with an arrow at the selection, clamped 12 px inside
the window, and above the first line when it does not fit below. It is
placed again on scroll and resize; when the selection leaves the window the
card closes as Esc does.

### 4. Sending goes through the chat

`EditorPage` passes a `send` from the overlay to `ChatPanel` as a request
(`{text, selection, id}`), which `ChatPanel` sends as if typed, with the card
in the stream. The panel stays mounted in every view, so this works while it
is hidden. A `draft` sets the panel's message box and quote.

### 5. Marked while open

While the card is open the selection is drawn with a second highlight
(`chatcms-selected`, the primary colour at 20 %), because focus moving into
the card clears the page's own selection.

## Risks / Trade-offs

- **Fonts**: the overlay cannot load Geist from another origin without CORS;
  it uses the system stack, which is close enough to read as the app and
  never as the site.
- **Older previews** that still have the previous overlay are replaced on
  the next load, as the script is served by the app.
