# Design

## Context

`BodyEditor` is a textarea with a toolbar that types Markdown marks. The page
editor hands it `loaded.body` and saves what `onChange` gives, through the
save queue that undo, the assistant and conflicts also go through. The body
is a string everywhere, and stays one: this change is inside the editor.

The serialisation spec already holds front matter to byte-for-byte writing
back; the body has had no such rule because a textarea never rewrites what
nobody typed. A formatted editor does.

## Goals / Non-Goals

**Goals:** formatted editing of the page text; a Markdown view; untouched
blocks written back byte for byte; Hugo shortcodes kept as locked blocks;
a safe way out for text the editor cannot hold.

**Non-Goals:** shortcodes declared in the model and edited through fields
(the next change); Quarto, Liquid and MDX dialects; the editor for front
matter fields of type `markdown`; collaborative editing.

## What the spike showed

Milkdown 7.23 with the CommonMark and GFM presets, run over our fixtures and
a file of tricky Markdown, keeps `_`/`*` emphasis, list markers, footnotes
and shortcode text as they are. Its whole-document writer does change:

| Source                         | Written back as                |
| ------------------------------ | ------------------------------ |
| A blank line after front matter | dropped                        |
| Hard break with two spaces     | `\` at the end of the line     |
| Ordered list `1. 2. 10.`       | renumbered `1. 2. 3.`          |
| `---` thematic break           | `***`                          |
| Setext heading                 | ATX `#` heading                |
| Reference link and definition  | inline link, definition gone   |
| `~~~` fence, indented code     | backtick fence                 |
| Table                          | padded and realigned           |
| `\<`                           | `<`                            |
| Body of only blank lines       | empty                          |

So the whole-document writer cannot be used for saving. The spike's
alternative, below, gave byte-identical output for every untouched block, and
a diff confined to the touched block for an edit, an insertion and a removal.

## Decisions

### 1. Keep the source; rewrite only touched blocks

On load the editor keeps the original text and the list of its top-level
blocks with their source ranges, from the same remark parser Milkdown uses
(`remarkCtx`, so the shortcode plugin applies to both), leaving out link
definitions, which have no node in the document. The document's top-level
nodes line up with those blocks one to one.

On a change, the old and the new top-level nodes are matched with a longest
common subsequence on `node.eq`. A matched node is written as its original
source slice, with the original separator before it when its neighbour is
matched too. An unmatched new node is written by Milkdown's serializer, on
its own, and joined with one blank line. The text before the first block and
after the last is kept. A link definition stays where it was as long as the
blocks around it are kept; a reference link inside an edited block is
written inline, which is correct Markdown and only touches that block.

### 2. Nothing is saved until somebody edits

`onChange` fires only from a transaction that changes the document and comes
from the person (not from loading or from a new value given from outside).
Switching views is not an edit.

### 3. A new value from outside rebuilds

When the body changes from outside (undo, the assistant, a change that
arrived), the editor takes the new text as its source and rebuilds, keeping
the cursor's block when it can. The editor never writes back text it did not
receive or edit.

### 4. Text that does not line up opens as Markdown

When the number of top-level nodes differs from the number of blocks (an
empty body counts as no blocks against Milkdown's one empty paragraph), the
editor opens the Markdown view with a note and does not offer the formatted
view for that text. Writing back from a mismatch could move text between
blocks, so it is not attempted.

### 5. The Hugo dialect is a remark transformer plus two nodes

A remark plugin walks the tree after parsing, with the source text from the
vfile. A paragraph or HTML block whose source starts a shortcode becomes a
`shortcode` node holding the exact source; for a paired shortcode, siblings
up to the matching closing tag are folded into it. A shortcode inside a line
becomes an inline `shortcode` node. In ProseMirror they are atom nodes, a
block and an inline one, with `raw` and `name` attributes, drawn as a locked
block or chip with the name. They are written back as raw `html` mdast nodes,
which remark writes verbatim. The plugin is the dialect: Quarto, Liquid and
MDX would each be another such plugin with the same two nodes.

### 6. The Markdown view is the textarea

Monospace, editable, showing exactly the saved text. Going back to the
formatted view rebuilds from that text with the rules above.

### 7. The toolbar uses editor commands

Heading, bold, italic, link, bullet list and image call Milkdown's commands.
Link and image ask for their address in a small popover with a field, not a
browser prompt. Labels come from the catalogue.

### 8. Loaded where it is used

The Milkdown editor lives in its own module, loaded with `React.lazy` by
`BodyEditor`, so screens without a page text do not carry it. While it
loads, the Markdown text shows read-only in its place.

### 9. Accessible and testable

The editable area carries `role="textbox"`, `aria-multiline="true"` and the
label "Page text", so screen readers, Testing Library and Playwright find it
by the same name as before. Tests that set the body type into the formatted
view or switch to the Markdown view where they need exact text.

## Risks / Trade-offs

- **Bundle size**: Milkdown, ProseMirror and remark add a few hundred
  kilobytes; loaded lazily and cached.
- **An edited block is normalised**: escapes, list numbering or emphasis
  marks in the touched block follow Milkdown's writer. The diff stays
  inside that block, which is the promise.
- **jsdom**: Milkdown runs under jsdom (the spike did), so Vitest can test
  the round trip; layout-dependent behaviour is left to e2e.
- **Paired shortcodes with Markdown inside** become one locked block in this
  change, so their inner text is edited in the Markdown view until the next
  change lets a declared shortcode open its fields.
