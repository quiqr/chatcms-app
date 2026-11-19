## MODIFIED Requirements

### Requirement: A page's text is written in Markdown either way

The text editor SHALL show the page text formatted, as it reads, and SHALL
write Markdown. The toolbar SHALL format what is selected, or the place
where the cursor is, without showing Markdown marks. The Markdown view SHALL
show exactly what will be saved and SHALL be editable.

#### Scenario: the toolbar

- **WHEN** somebody selects a word and presses Bold
- **THEN** the word SHALL show in bold, without asterisks
- **AND** the saved text SHALL hold the word in Markdown bold

#### Scenario: the source

- **WHEN** somebody asks for the source
- **THEN** exactly what will be saved SHALL be shown
- **AND** the toolbar SHALL be put away

#### Scenario: back from the source

- **WHEN** somebody changes the text in the Markdown view and goes back to
  the formatted view
- **THEN** the formatted view SHALL show the changed text
