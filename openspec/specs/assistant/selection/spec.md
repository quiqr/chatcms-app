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
the selected text, its context, and where the text comes from: a field of a
page, shared text, or the site's design. Where the text comes from SHALL be
found in the site's source as the model describes it, for any generator,
starting with the page the preview address maps to where that is known (for
a Hugo site, through Hugo's own list of pages and their addresses). When
nothing can be worked out, the assistant SHALL be told to find the text
through its page tools.

#### Scenario: a page that matches

- **WHEN** the person selects the title on the preview of the home page and
  asks for a change
- **THEN** the assistant SHALL be told it is the field "title" of the page
  "home", and its change SHALL be saved as the person's change to that page

#### Scenario: shared text

- **WHEN** the person selects a footer text held by a single in
  `data/settings.yaml`
- **THEN** the assistant SHALL be told it is shared text that shows on more
  pages, and to say so before changing it

#### Scenario: the design

- **WHEN** the person selects text no page or shared file holds
- **THEN** the assistant SHALL be told it comes from the site's design, which
  it cannot change

#### Scenario: no page matches

- **WHEN** where the text comes from cannot be worked out at all
- **THEN** the assistant SHALL be told no page was found for the address

### Requirement: A sent request shows its progress at the selection

After a request is sent from the selection, a pill SHALL stand where the
input was, following the text it is about. While the request is being
answered it SHALL say it is working and offer Stop. When the answer changed
the page it SHALL say "Changed" with Undo and Show in chat, and the changed
text SHALL be tinted for a moment; the pill SHALL stay at least 8 seconds and
until the person touches the preview again. When the answer is a question it
SHALL say so and offer Answer in chat. Every request SHALL have a pill of its
own, and pills SHALL survive a reload of the preview.

#### Scenario: working, then done

- **WHEN** somebody sends "Make it shorter" from a selection
- **THEN** a pill says "Working on it…" with Stop
- **AND** once the page is changed it says "Changed" with Undo and Show in
  chat, and the new text is tinted

#### Scenario: undo at the selection

- **WHEN** somebody presses Undo on that pill
- **THEN** the page holds what it held before, the chat says the change was
  undone, and the pill goes

#### Scenario: a question

- **WHEN** the assistant answers with a question and changes nothing
- **THEN** the pill says the assistant has a question, and Answer in chat
  opens Chat + preview

#### Scenario: two at once

- **WHEN** somebody sends two requests from two selections
- **THEN** each has its own pill, and Chat + preview counts 2

#### Scenario: stop

- **WHEN** somebody presses Stop on a working pill
- **THEN** the assistant stops answering and the pill goes

### Requirement: Where a selection comes from can be asked

The application SHALL answer, for a selection, whether its text is in a
field of a page (with the page, its title and the field), shared text (with
the single and the field) or part of the site's design, to anybody who may
use the chat in that project.

#### Scenario: asking

- **WHEN** the editor asks where "Korenaar Bakery" on the home page comes from
- **THEN** the answer is the field "title" of "Home page"
