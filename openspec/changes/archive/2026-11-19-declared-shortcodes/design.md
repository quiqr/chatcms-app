# Design

## Context

`markdown-editor` turned shortcodes into locked atom nodes (`shortcode_block`,
`shortcode_inline`) holding their exact source in `raw`, through a remark
plugin that is the Hugo dialect. The page editor reads a page as controls
from the server (`internal/forms`), built from the model (`internal/model`),
which merges `base.yaml` with `includes/` and `_mergePartial`. Picture fields
use a picker that the media library provides through `withPicking`.

## Goals / Non-Goals

**Goals:** shortcode declarations in the model; a popup form for a declared
shortcode; minimal rewriting of its parameters; inserting from the toolbar;
a block that shows what it holds.

**Non-Goals:** list and group fields in shortcodes; a formatted editor for
the inner text; Quarto, Liquid and MDX; finding pictures in page bundles.

## Decisions

### 0. Syntax lives in a dialect, chosen by runtime

The model declares what a shortcode holds; it never names delimiters.
Delimiters alone cannot describe a generator's syntax (Hugo's `/name`
closing tag, Liquid's `endname`, MDX's JSX attributes and expressions,
quoting rules), so a dialect in code reads and writes it: the remark
transformer, the parameter tokenizer and the tag builder. The server
resolves the dialect for a page: `shortcodeDialect` from the base when
given, otherwise by runtime (Hugo, or no runtime, gives `hugo`; any other
runtime gives none in this change). The page read returns it; the editor
loads that dialect's remark plugin, or none. `inner` carries the meaning,
and the Hugo dialect turns `inner: markdown` into `{{% %}}` and anything
else into `{{< >}}`, which is what Hugo means by the two.

### 1. Declarations are model data, like singles

`internal/model` reads `shortcodes` from the base and from
`includes/shortcodes/*`, merged by key the way singles and collections are,
`_mergePartial` included. Each declaration's fields go through the existing
field decoding, then a check: single-value types only, positions unique and
from 0 without gaps, `inner` empty, `markdown` or `text`. The model's golden
JSON gains the list; existing goldens are unchanged because the key is new.

### 2. Declarations travel with the page

The page read already returns controls; it gains `dialect` and
`shortcodes`: for each declaration its key, title, inner and its fields as
controls without values. The frontend fills the values from the parsed shortcode. One source
for field rendering, `Fields.tsx`, serves both the page form and the popup.

### 3. Parameters are parsed into spans

A small tokenizer reads a shortcode's opening tag into spans: the name, then
each parameter as `{name?, value, quote, from, to}` over the raw text,
positional ones numbered in order. Reading gives field values. Writing a
changed value replaces only that span's value, keeping its quote style (and
adding quotes when the new value needs them). A value newly given gets a new
span appended at the end of the tag; a value cleared removes its span and
the space before it. The inner text is the slice between the opening and
the closing tag, replaced whole when it changed.

### 4. Declared blocks are a node view

The two nodes keep `raw` and `name`. When the editor knows a declaration for
`name`, a ProseMirror node view draws the block with its title, the first
text value and a picture thumbnail, and opens the form on click or Enter.
Done replaces the node with one whose `raw` is the rewritten source, in one
transaction, so it is one step for undo and one save, and the source keeper
writes only that block back.

### 5. The form is a dialog with the page's field controls

A dialog (popover on desktop, sheet on a phone) holds `Fields` for the
declaration's controls and, with `inner`, a Markdown textarea. Values are
local until Done; Cancel changes nothing.

### 6. Inserting builds a new tag

The toolbar's Insert menu lists the declarations. Done asks the dialect to
build the tag; for Hugo: delimiter by `inner`, positional values in position order, then named
values in model order, empty ones left out, values quoted with `"` when they
need quoting; with `inner`, the inner text and the closing tag on their own
lines. It inserts a `shortcode_block` after the cursor's block.

## Risks / Trade-offs

- **Hugo's parameter syntax** has raw strings (backticks) and escaped quotes;
  the tokenizer handles `"…"`, `` `…` `` and bare words, and refuses (leaving
  the block locked) a tag it cannot read, rather than guessing.
- **A shortcode used inline** (in running text) is edited through the same
  form from its chip; inserting is block-only.
- **Pictures in page bundles** are not found by the library; unchanged by
  this change.
