# content/markdown-text Specification

## Purpose
TBD - created by archiving change markdown-editor. Update Purpose after archive.

## Requirements

### Requirement: Opening a page writes nothing

Opening a page and looking at its text in either view SHALL NOT save the
page.

#### Scenario: Just reading

- **WHEN** somebody opens a page, switches between the formatted and the
  Markdown view and leaves
- **THEN** no save is made and the file is unchanged

### Requirement: Untouched text is written back byte for byte

When the text is saved after an edit in the formatted view, every block the
edit did not touch SHALL be written back exactly as it was, and so SHALL the
text before the first block, between untouched blocks, and after the last
block.

#### Scenario: One paragraph changed

- **WHEN** a page holds a setext heading, a reference link with its
  definition, an ordered list starting at 10, a tilde fence and a table, and
  somebody changes one word in an unrelated paragraph
- **THEN** the saved text differs from the old only in that paragraph

#### Scenario: A block added and a block removed

- **WHEN** somebody adds a paragraph between two others and removes a third
- **THEN** the saved text holds the new paragraph separated by one blank line,
  lacks the removed one, and is otherwise unchanged

### Requirement: A shortcode is a locked block

A Hugo shortcode the model does not declare SHALL show in the formatted view
as a locked block naming the shortcode, or as a locked chip when it sits
inside a line of text. It SHALL be possible to move or remove it, not to
change what it says, and it SHALL be written back byte for byte. A paired
shortcode, with its closing tag and everything between, SHALL be one block.

#### Scenario: A figure

- **WHEN** a page holds `{{< figure src="/images/loaf.jpg" caption="Our loaf" >}}`
  and the model declares no `figure`
- **THEN** the formatted view shows a locked block named "figure"
- **AND** after an edit elsewhere the line is saved as it was

#### Scenario: A notice around text

- **WHEN** a page holds `{{% notice info %}}`, a line of text and
  `{{% /notice %}}`, and the model declares no `notice`
- **THEN** the three lines are one locked block named "notice"

### Requirement: Text the editor cannot hold is opened as Markdown

When the formatted view cannot match its blocks one to one with the text, the
editor SHALL open the text in the Markdown view, SHALL say that this text can
only be edited as Markdown, and SHALL NOT offer the formatted view for it.

#### Scenario: Blocks that do not line up

- **WHEN** a page's text cannot be matched block for block
- **THEN** the Markdown view opens with a note, and saving writes what is in
  it

### Requirement: Shortcodes are read in the site's dialect

The editor SHALL read and write shortcodes in one dialect: the one the model
names with `shortcodeDialect`, otherwise the one that belongs to the site's
runtime. Hugo sites, and sites without a runtime, SHALL use the Hugo
dialect. Quarto sites SHALL use the Quarto dialect, which locks `{{< >}}`
shortcodes and offers no form. A site whose runtime has no dialect in this
version SHALL have no shortcodes recognised.

#### Scenario: A Hugo site

- **WHEN** a Hugo site's page holds `{{< figure src="a.jpg" >}}`
- **THEN** it is a shortcode block

#### Scenario: A Quarto site

- **WHEN** a Quarto site's page holds `{{< video https://example.com/a.mp4 >}}`
  and `{{% not quarto %}}`, and the paragraph around the first is edited
- **THEN** the video shortcode is a locked chip and is written back as it
  was, link included, and the second is plain text

### Requirement: A declared shortcode is edited through its form

A shortcode the model declares SHALL show as a block with its title, its
first text value and, when it has a picture field with a value, a small
picture. Choosing it SHALL open a form with its fields, the controls forms
use, and a text box for what is between its tags when it declares `inner`.
Done SHALL write the shortcode back, changing only the values that changed:
every other parameter, undeclared ones included, SHALL keep its place, its
quotes and its spelling.

#### Scenario: Changing a caption

- **WHEN** somebody opens `{{< figure src="/images/loaf.jpg" class="wide" caption="Our loaf" >}}`
  and changes the caption to "Fresh loaf"
- **THEN** the text holds `{{< figure src="/images/loaf.jpg" class="wide" caption="Fresh loaf" >}}`

#### Scenario: A positional parameter and the text between

- **WHEN** somebody opens `{{% notice info %}}` … `{{% /notice %}}`, chooses
  warning and changes the text between
- **THEN** the text holds `{{% notice warning %}}`, the new text and
  `{{% /notice %}}`

#### Scenario: A picture

- **WHEN** somebody chooses a picture for the figure through the library
- **THEN** its address is written as `src`, and the library lists the page
  as using it

### Requirement: A declared shortcode can be inserted

The formatted view's toolbar SHALL offer the declared shortcodes. Choosing
one SHALL open its form, and Done SHALL insert it as a block where the
cursor is, written by the dialect: positional parameters first in order and
named ones after in the order the model declares them, empty ones left out.
The Hugo dialect SHALL write `{{% %}}` for `inner: markdown` and `{{< >}}`
otherwise.

#### Scenario: Inserting a figure

- **WHEN** somebody inserts a figure with a picture and a caption
- **THEN** the text holds `{{< figure src="…" caption="…" >}}` as a block of
  its own

#### Scenario: No declarations

- **WHEN** the model declares no shortcodes
- **THEN** the toolbar offers no insert menu
