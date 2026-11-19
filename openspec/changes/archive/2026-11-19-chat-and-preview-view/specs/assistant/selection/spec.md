## MODIFIED Requirements

### Requirement: A selection is quoted into the chat

A selection from the preview SHALL open Chat + preview with the selection
quoted above the message box. The person MAY remove the quote; a message
sent with a quote SHALL carry the selection with it.

#### Scenario: quoting

- **WHEN** a selection arrives from the preview
- **THEN** the editor SHALL switch to Chat + preview with the selected text
  quoted

#### Scenario: dropping the quote

- **WHEN** the person removes the quote and sends a message
- **THEN** the message SHALL carry no selection
