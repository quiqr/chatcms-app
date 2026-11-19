# Brief for Claude Design: the chat interface in the editor

Quiqr ChatCMS lets non-technical people edit a website through forms, a live
preview and an AI assistant. The assistant already works. This brief asks for
the design of how a person talks to it, in particular from the preview itself.

## Who uses it

Content creators who are not technical. They never see git or code words:
no commit, branch, merge, deploy. They see "saved", "changes", "publish",
"undo". Implementers see the same screens with technical details folded away.

## What exists today

Design system, tokens and screens: `docs/design/DESIGN.md` (tokens in
`docs/design/tokens/`). Use those tokens and the shadcn components named
there, not new colours.

- **The editor** has a top bar, the project navigation, the content menu, and
  a row of slots: the form, the preview and the assistant. A switch offers
  Form, Side by side and Preview.
- **The page never scrolls.** Each panel scrolls on its own, and the header
  always stays.
- **On a window too narrow for form and preview side by side**, the editor
  shows one of them, the form first.
- **Where form, preview and assistant do not all fit**, the assistant takes
  the form's place beside the preview.
- **The assistant panel** is a chat column of 340 px: the conversation, a
  message box at the bottom that always stays in view, and for every change
  the AI made a line "Changed <page>" with Show and Undo.
- **Selecting text in the preview** shows a small round button next to the
  selection. Pressing it opens the assistant with the selection quoted above
  the message box; the person types what should change. After the AI's
  change the preview reloads and the new text is highlighted for four
  seconds.
- **On a phone** the editor has Edit and Preview tabs and an Assistant button
  in a bar at the bottom; the assistant then fills the screen.
- **A PDF preview** shows the document's pages at their own size, and text in
  it can be selected the same way.

## What to design

### 1. A small input over the preview

Selecting text in the preview opens a small input right at the selection
instead of jumping to the chat. When the person sends it, the request is
added to their chat stream. This must work in every view that shows a
preview, also when the chat column is closed.

Questions to answer: where the input sits relative to the selection and the
preview's edges, what happens when the preview scrolls underneath it, how it
is dismissed, and what it looks like on a phone.

### 2. Progress and Undo at the selection

After sending, the person sees the request's state next to the selection
itself: "working on it", then "done", with an Undo right there. When the AI
asks something back, a badge near the selection points to the chat.

Questions: how long "done" and Undo stay, what happens when the person keeps
working elsewhere, and how several requests at once look.

### 3. A Chat and preview view

The view switch becomes Form, Form and preview, Chat and preview, Preview.
Chat and preview is for working only through the assistant.

Questions: how the four views read as one switch, what the narrow window
shows for Chat and preview, and whether the assistant button in the top bar
and this view are one control or two.

### 4. A canvas request in the chat stream

A request made on the preview shows in the chat stream as a card with a small
picture of what was selected, rather than as quoted text. The AI's changes
keep their summary with Show changes and Undo.

## Constraints

- Desktop from 640 px up and the phone frame below that; check both.
- Touch targets of at least 44 px, no sideways scrolling on a phone.
- Every string comes from the copy catalogue; no git words for editors.
- The preview may be on another origin than the app. Anything drawn inside
  the preview (the input, the badges) is drawn by a small script in the
  site's page, so keep it self-contained, small, and visually separate from
  the site's own styles.
- The chat stream is private per person per project.

## What we would like back

The same kind of hand-off as `docs/design/folded-menu/`: a spec in Markdown
with numbered sections, reference screenshots and static HTML for each state
at desktop and phone width, and a short prompt for Claude Code that says what
to build and how to check it.
