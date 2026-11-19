---
# chatcms-oren
title: 'A real Markdown editor: Milkdown with shortcode blocks'
status: in-progress
type: epic
priority: high
created_at: 2026-11-18T12:37:26Z
updated_at: 2026-11-19T20:00:00Z
openspec-link: openspec/changes/archive/2026-11-19-markdown-editor
parent: chatcms-iqog
---

## Purpose

Replace the hand-written textarea in `web/src/fields/BodyEditor.tsx` with a
real WYSIWYG editor that saves clean Markdown. Decision taken 18 November
2026: **Milkdown**, no TipTap proof of concept. Markdown is Milkdown's native
document format, which gives round-trip fidelity, and remark plugins make a
shortcode dialect per static site generator possible without rebuilding the
syntax layer.

## Current state

- `web/src/fields/BodyEditor.tsx` is a plain React textarea.
- "Formatted" and "markdown" modes are the same textarea; only the font
  differs. The toolbar inserts Markdown symbols, so it is not WYSIWYG.

## Scope

- Replace BodyEditor with Milkdown plus a raw Markdown mode.
- Round-trip fidelity is a hard requirement: untouched files are not
  reformatted, git diffs stay minimal, front matter and shortcodes survive.
- Shortcodes are declared in the Quiqr model, like forms: fields per
  shortcode (e.g. figure: image, caption, alt). In the editor a shortcode is
  a block with a preview; clicking it opens the familiar form fields.
- Unknown shortcodes become locked raw blocks so they cannot break.
- A dialect per runtime: parser and serializer plugins, with the model
  declarations staying generic. Different SSGs have different syntax (Hugo
  `{{< >}}` / `{{% %}}`, Quarto `{{< >}}`, Jekyll/Eleventy Liquid `{% %}`,
  MDX components).
- Hugo dialect first, designed so Quarto, Liquid and MDX can follow.

## Out of scope (follow-up)

- Quarto dialect: callouts and fenced divs (`:::`), code cells with rendered
  output, inline code chips, cross-references, and includes shown as
  collapsible read-only blocks.
- Shared includes: reverse index of documents per include, re-render on
  publish, role-restricted editing.

## Acceptance

- A round-trip test suite on tricky Markdown and shortcode fixtures shows
  zero diff for untouched content.
- Unknown shortcodes round trip byte for byte as locked blocks.
- A declared shortcode can be edited through its model fields and serialises
  back to the correct dialect syntax.

## References

- `web/src/fields/BodyEditor.tsx`, `web/src/routes/PageEditor.tsx`
- `docs/design/DESIGN.md` section 5.6
- Source notes: `content/quiqr_chatcms__markdown_editor.md` in the second
  brain, decision of 18 November 2026.
