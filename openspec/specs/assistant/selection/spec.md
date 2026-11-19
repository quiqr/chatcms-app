# assistant/selection Specification

## Purpose
Defines asking the assistant about a selection in the preview, and how the
page it came from is found.

## Requirements

### Requirement: A selection is quoted into the chat

A finished selection of two characters or more in the preview SHALL show a
chip at the end of the selection for editors and implementers, in Preview
and Chat + preview. Choosing the chip SHALL open an input at the selection.
Sending from it SHALL add the request, with the selection, to the chat and
ask the assistant, without changing the view. Closing the input SHALL keep
what was typed as a draft in the chat's message box with the selection
quoted. The person MAY remove the quote there; a message sent with a quote
SHALL carry the selection with it.

#### Scenario: the chip waits for the selection to end

- **WHEN** somebody drags across a sentence
- **THEN** no chip shows while they drag, and the chip shows once they let go

#### Scenario: too short, or not allowed

- **WHEN** a single character is selected, or a viewer selects text, or the
  editor shows Form + preview
- **THEN** no chip shows

#### Scenario: sending from the selection

- **WHEN** somebody opens the input, types "Make it shorter" and presses Enter
- **THEN** the request is in the chat with the selection, the assistant is
  asked, and the view stays as it was

#### Scenario: a quick action

- **WHEN** somebody chooses "Fix spelling" in the input
- **THEN** that request is sent at once

#### Scenario: quoting

- **WHEN** somebody types in the input and presses Esc
- **THEN** the input closes and the chat's message box holds the text, with
  the selection quoted

#### Scenario: no room below

- **WHEN** the selection is near the bottom of the preview
- **THEN** the input opens above it, inside the preview

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
