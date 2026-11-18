# Proposal

Bean: [chatcms-5odk](../../../.beans/chatcms-5odk--select-to-edit-in-the-preview.md)
(select-to-edit in the preview), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

Asking the assistant for a change means naming the page and quoting the
text by hand. In the preview the person can already see the sentence they
want changed; selecting it should be enough to say which one.

## What Changes

- **A selection script in every web preview.** The preview proxy, on the
  app's address and on the preview origin, adds one script to each HTML
  page it passes through. On a selection it shows a small button; pressing
  it posts the selection, the text around it and the page's address to the
  app, by origin-checked `postMessage`.
- **The selection is quoted into the chat panel.** The panel opens with the
  quote above the message box; the person types what should change, and
  can drop the quote.
- **The address is turned into a page.** For a Hugo site, `hugo list all`
  over the person's copy maps the preview address to a content file, and
  the model's folders map that to the page key the assistant's tools use.
  The assistant is told the page, the selection and its context; when no
  page matches it is told to look for the text through its tools.
- **The changed text is highlighted.** After the assistant's change the
  preview reloads, and the text between the selection's old context is
  highlighted for a few seconds. Show and Undo stay on the change, as now.
- **PDF previews move to PDF.js**, served from the preview's own address so
  the PDF never renders beside the app. It keeps the page position on
  refresh and allows selection, which goes to the assistant the same way,
  to be found in the sources.

## Capabilities

### New Capabilities

- `preview/selection`: the script in a preview, what it may send and
  receive, and the highlight after a change.
- `assistant/selection`: asking about a selection, and how its page is found.

### Modified Capabilities

- `preview/pane`: a pdf preview is shown by PDF.js, with selection.

## Impact

- `internal/preview`: the proxy injects the script and serves
  `/__chatcms/` (the script, the PDF viewer) under each preview's address.
- `internal/assistant`: `Ask` takes an optional selection; a `Locator`
  turns an address into a page.
- `internal/model`: `KeyOf(file)`, the content browser's key for a file,
  shared with `sitesync`.
- Frontend: the pane listens for selections; the chat panel holds a quote;
  the editor shell's assistant can be opened from outside. A second, small
  Vite build for the PDF viewer, embedded with the app.
- Acceptance (milestone 07, epic 8): selecting a sentence on a Hugo page and
  asking for a change edits the right page, highlights the change and saves
  it as the person's change.

## Assumptions

- **Quote into chat, not an inline prompt.** The bean allows either; one
  place to talk to the assistant is simpler to learn and to test, and the
  phone has no room for a popover over the preview.
- **The script is served, not inlined**, at `<preview>/__chatcms/select.js`,
  so a site with a strict content security policy for inline scripts still
  gets it. The app's origin is given to it in a data attribute.
- **No mapping is cached.** `hugo list all` runs when a question with a
  selection is asked, which is seldom and takes well under a second; there
  is then nothing to refresh on change.
- **Text that is not on a page is reported, not changed.** The assistant
  edits only through the page tools; when the selected text comes from a
  template or data file, it says so instead of searching files it may not
  change.
- **"A commit on the user's branch"** is the person's save in their own
  copy, which every assistant change already is.
- **The PDF viewer lives on the preview's address.** Rendering a site's PDF
  with PDF.js inside the app would put a parser of untrusted files beside
  the app's session; on the preview's address it is as far away as the
  site's own scripts. It posts selections exactly like the web script.
- **No highlight in a PDF.** The PDF reloads at the same page; finding the
  changed words in a rendered PDF is left to the Quarto epic.
- **The script inside a site keeps its own colours.** `check:interface`
  skips the colour rule under `web/src/preview/`: that code runs in the
  customer's page, where the app's tokens do not exist.
- **The chat panel waits for the assistant's stream** before it sends,
  holding the message for up to two minutes. Opening the panel by a
  selection sends right away, and a reply could finish before the stream
  was there to carry it.
