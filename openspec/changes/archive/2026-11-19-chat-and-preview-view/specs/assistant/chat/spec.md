## MODIFIED Requirements

### Requirement: What the agent changed is listed with Show and Undo

Every page the agent changed while answering SHALL be listed in the chat as
a change summary, "Changed <page>", with Show changes and Undo. Show changes
SHALL open that page and, when the request came from a selection, mark the
changed place in the preview. Undo SHALL go back to the version before the
change, as a new version, without publishing.

#### Scenario: a change

- **WHEN** the agent changes the home page's title
- **THEN** the chat SHALL list "Changed Home page" with Show changes and Undo
- **AND** the preview SHALL show the new title

#### Scenario: undo

- **WHEN** the person presses Undo on that change
- **THEN** the page SHALL hold what it held before, as a new version
- **AND** nothing SHALL be published

## ADDED Requirements

### Requirement: A request from the preview is a card

A message sent with a selection SHALL show in the chat as a card: a snippet
of the page with the selection marked, the message, and "From the preview ·
<page>". Choosing the card SHALL scroll the preview to the selection and
mark it.

#### Scenario: a request card

- **WHEN** somebody selects a sentence on the news page and asks to shorten it
- **THEN** the chat shows a card with the sentence marked in its context, the
  request, and "From the preview · News"

#### Scenario: back to the selection

- **WHEN** somebody chooses that card
- **THEN** the preview scrolls to the sentence and marks it

### Requirement: Running requests are counted on the view switch

While a request is being worked on and the chat is not shown, the Chat +
preview segment SHALL carry a badge with the number of requests still
running. There SHALL be no separate Assistant button on the desktop.

#### Scenario: working in the form

- **WHEN** a request is running and somebody switches to Form
- **THEN** Chat + preview carries a badge "1" until the request is done
