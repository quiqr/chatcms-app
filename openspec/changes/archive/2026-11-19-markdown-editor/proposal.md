# Proposal

Bean: [chatcms-oren](../../../.beans/chatcms-oren--a-real-markdown-editor-milkdown-with-shortcode-bloc.md)
(a real Markdown editor: Milkdown with shortcode blocks), milestone 03.1
([chatcms-iqog](../../../.beans/chatcms-iqog--ui-cleanups.md)).

## Why

The page text is a plain textarea. "Formatted" and "Markdown" differ only in
font, and the toolbar types Markdown marks into the text, so a content creator
reads `**bold**` and `## Heading` instead of bold and a heading. Content
creators should edit text as it looks, while the file in git stays exactly as
the site's author wrote it: an edit to one paragraph must not reformat the
rest of the page.

## What Changes

- **A formatted editor (Milkdown)** replaces the textarea for the page text:
  headings, bold, italic, links, lists and images show as they look, and the
  toolbar formats instead of typing marks.
- **The Markdown view** stays, showing exactly what will be saved, and can be
  edited directly.
- **The text keeps its own writing.** Opening a page saves nothing. An edit
  rewrites only the blocks it touched (a paragraph, a list, a heading); every
  other block, the blank lines between them and the start and end of the text
  are written back byte for byte.
- **Shortcodes are locked blocks.** A Hugo shortcode (`{{< >}}`, `{{% %}}`,
  with or without a closing tag) shows as a block naming the shortcode, that
  can be moved or removed but not changed in the formatted view; it is
  written back byte for byte. An inline shortcode is a locked chip.
- **Text the editor cannot hold safely** (where its blocks do not line up
  with the source) opens in the Markdown view, saying why, and the formatted
  view is not offered for it.

## Capabilities

### New Capabilities

- `content/markdown-text`: how the page text is edited and written back.

### Modified Capabilities

- `interface/form-fields`: the page text is edited formatted, with a
  Markdown view.

## Impact

- `web/src/fields/BodyEditor.tsx` is rewritten around Milkdown; new modules
  for keeping the source and for the shortcode dialect.
- `web/package.json` gains `@milkdown/kit`; `flake.nix`'s `npmDepsHash`
  changes with it. The editor's code is loaded only where it is used.
- `PageEditor`, saving, undo and the assistant keep working with the body as
  a string; no Go code changes.

## Assumptions

- **Split.** The bean is done in two changes. This one: the formatted editor,
  faithful writing back, unknown shortcodes as locked blocks, the Hugo
  dialect. The next: shortcodes declared in the Quiqr model and edited
  through their fields.
- **Measured, not assumed.** A spike round-tripped our fixtures and a file of
  tricky Markdown through Milkdown 7.23 with CommonMark and GFM. Untouched
  structure is mostly kept, but the whole-document writer changes a leading
  blank line, hard breaks, ordered list numbers, thematic breaks, setext
  headings, reference links, tilde and indented code, table padding and some
  escapes. Writing back only the touched blocks gave a byte-identical result
  for every untouched block. See `design.md`.
- **Fields of type `markdown`** in the front matter keep their current
  control; giving them this editor is a follow-up.
- **Quarto, Liquid and MDX dialects** are out of scope; the dialect is one
  module so they can follow.
