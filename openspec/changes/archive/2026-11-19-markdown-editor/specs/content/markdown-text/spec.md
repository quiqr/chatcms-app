## ADDED Requirements

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

A Hugo shortcode SHALL show in the formatted view as a locked block naming
the shortcode, or as a locked chip when it sits inside a line of text. It
SHALL be possible to move or remove it, not to change what it says, and it
SHALL be written back byte for byte. A paired shortcode, with its closing
tag and everything between, SHALL be one block.

#### Scenario: A figure

- **WHEN** a page holds `{{< figure src="/images/loaf.jpg" caption="Our loaf" >}}`
- **THEN** the formatted view shows a locked block named "figure"
- **AND** after an edit elsewhere the line is saved as it was

#### Scenario: A notice around text

- **WHEN** a page holds `{{% notice info %}}`, a line of text and
  `{{% /notice %}}`
- **THEN** the three lines are one locked block named "notice"

### Requirement: Text the editor cannot hold is opened as Markdown

When the formatted view cannot match its blocks one to one with the text, the
editor SHALL open the text in the Markdown view, SHALL say that this text can
only be edited as Markdown, and SHALL NOT offer the formatted view for it.

#### Scenario: Blocks that do not line up

- **WHEN** a page's text cannot be matched block for block
- **THEN** the Markdown view opens with a note, and saving writes what is in
  it
