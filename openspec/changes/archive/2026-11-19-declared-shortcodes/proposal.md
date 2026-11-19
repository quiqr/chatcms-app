# Proposal

Bean: [chatcms-oren](../../../.beans/chatcms-oren--a-real-markdown-editor-milkdown-with-shortcode-blo.md)
(a real Markdown editor: Milkdown with shortcode blocks), second change after
`markdown-editor`. Milestone 03.1
([chatcms-iqog](../../../.beans/chatcms-iqog--ui-cleanups.md)).

## Why

Since `markdown-editor`, every Hugo shortcode in the page text is a locked
block: safe, but a content creator cannot change a figure's caption or a
notice's text without the Markdown view. Shortcodes are how Hugo sites put
structured pieces in running text, so they should be edited the way the
rest of a page is: through fields the site's model declares.

## What Changes

- **The model declares shortcodes** under a new `shortcodes:` key in
  `quiqr/model/base.yaml`, and in `quiqr/model/includes/shortcodes/`, like
  singles and collections. Each has a key (its name in the text), a title
  and fields of the kinds forms already have. A field with `position: n` is
  the n-th positional parameter; without it, a named one. `inner: markdown`
  or `inner: text` says the shortcode has a closing tag and what is between.
- **A declared shortcode opens a popup form** when clicked: its fields with
  the controls forms use, the picture field included (choose or upload
  through the media library), and a text box for what is between the tags.
- **Writing back changes only what changed**: a changed parameter's value is
  replaced where it stands; other parameters, their order, their quotes and
  undeclared parameters are kept.
- **Inserting a shortcode** from the toolbar: a menu of the declared
  shortcodes, which opens the same form and inserts the shortcode on Done.
- **The syntax belongs to a dialect, not to the model.** The declarations
  say what a shortcode holds; a dialect in code reads and writes the
  generator's syntax. The dialect follows the site's runtime (Hugo by
  default), and `shortcodeDialect:` in the base overrides it. This change
  brings the Hugo dialect, and a Quarto dialect that only locks `{{< >}}`
  shortcodes; on a runtime without a dialect, shortcodes are not
  recognised, as their syntax is unknown. In Hugo,
  `inner: markdown` is written `{{% %}}` and `inner: text` `{{< >}}`, which is
  what Hugo means by the two.
- **The block shows what it holds**: the shortcode's title, its first text
  value, and a small picture when it has a picture field with a value.
- An undeclared shortcode stays a locked block, as now.
- A Quarto site no longer has Hugo's syntax applied to its text: its
  `{{< >}}` shortcodes are locked, and `{{% %}}` is plain text.
- A shortcode inside a line is found in the line's source, so one holding a
  link or other markup stays whole when the line is edited.

## Capabilities

### New Capabilities

### Modified Capabilities

- `content/model-schema`: the model declares shortcodes.
- `content/markdown-text`: declared shortcodes are edited through a form and
  inserted from the toolbar.

## Impact

- `internal/model`: reading `shortcodes` from the base and from includes;
  goldens for a fixture with shortcodes.
- `internal/httpapi` (page read): the declarations travel with the page, as
  field descriptions without values.
- `web/src/fields/markdown/`: parameter parsing and rewriting, the popup
  form, the insert menu, the block's summary.
- The old Quiqr desktop app ignores the new key, so a model stays readable
  by both.

## Assumptions

- Declared through the bean's discussion of 19 November 2026: place A1
  (`shortcodes:` in the model), popup form, a Markdown text box for the
  inner text (not a nested formatted editor), and the syntax in a dialect
  chosen by runtime with an override in the model, rather than delimiters
  in the model: delimiters alone cannot describe closing tags, quoting or
  positional parameters.
- A picture uploaded from the popup goes to the site's shared picture folder,
  as a picture field in a form does; the library finds its use because it
  reads the whole file.
- Leaving Quarto's shortcodes as plain text, as first agreed, turned out to
  break them: editing the line around `{{< video https://… >}}` turns the
  address into an autolink. Quarto uses the same `{{< >}}`, so its dialect
  locks them.
- The thumbnail of a picture in a block, like the picture field's in a form,
  uses the site address, which the app does not serve; a picture that does
  not load is left out. Filed as a bean of its own.
- Field types allowed in a shortcode: those with a single value (string,
  text, number, boolean, select, date, picture). Lists and nested groups are
  out of scope.
- Quarto, Liquid and MDX dialects stay out of scope.
