# assistant/selection Specification

## Purpose
Defines asking the assistant about a selection in the preview, and how the
page it came from is found.

## Requirements

### Requirement: A selection is quoted into the chat

A selection from the preview SHALL open the assistant with the selection
quoted above the message box. The person MAY remove the quote; a message
sent with a quote SHALL carry the selection with it.

#### Scenario: quoting

- **WHEN** a selection arrives from the preview
- **THEN** the assistant SHALL open with the selected text quoted

#### Scenario: dropping the quote

- **WHEN** the person removes the quote and sends a message
- **THEN** the message SHALL carry no selection

### Requirement: The assistant is told which page a selection came from

For a message with a selection, the assistant SHALL be given the request,
the selected text, its context and, when the preview address maps to a page
of the model, that page's key. For a Hugo site the address SHALL be mapped
through Hugo's own list of pages and their addresses. When no page matches,
the assistant SHALL be told to find the text through its page tools, and to
say so when it is on no page.

#### Scenario: a page that matches

- **WHEN** the person selects text on the preview of a page of the model and
  asks for a change
- **THEN** the assistant SHALL be given that page's key, and its change
  SHALL be saved as the person's change to that page

#### Scenario: no page matches

- **WHEN** the selection is on an address no content file is served at
- **THEN** the assistant SHALL be told no page was found for the address
