## MODIFIED Requirements

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
