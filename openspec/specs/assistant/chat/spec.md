# assistant/chat Specification

## Purpose
Defines talking to the agent from the editor: sending, the reply as it is
written, and what the agent changed.

## Requirements

### Requirement: A message gets a streamed reply

A person SHALL be able to send a message from the editor's assistant panel to
the agent of their workspace, and SHALL see the reply as it is written.

#### Scenario: asking

- **WHEN** a person sends "change the title of the home page to Korenaar by
  night"
- **THEN** the reply SHALL appear in the panel piece by piece as it is written

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

### Requirement: Without an assistant, the panel says so

When the installation has no gateway or no agent, the panel SHALL say the
assistant is not available rather than fail.

#### Scenario: no gateway

- **WHEN** a person opens the panel on an installation without a gateway
- **THEN** it SHALL say the assistant is not available

### Requirement: The panel speaks without git words

What the panel shows a content creator SHALL use none of the words the copy
rules keep from them.

#### Scenario: the words

- **WHEN** the panel lists a change
- **THEN** it SHALL contain none of the forbidden words

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

While a request is being worked on, or a question from the assistant is not
yet answered, and the chat is not shown, the Chat + preview segment SHALL
carry a badge with their number. There SHALL be no separate Assistant button
on the desktop.

#### Scenario: working in the form

- **WHEN** a request is running and somebody switches to Form
- **THEN** Chat + preview carries a badge "1" until the request is done

#### Scenario: an open question

- **WHEN** the assistant asked a question and somebody is in Preview
- **THEN** Chat + preview carries a badge until they send their answer

### Requirement: A running answer can be stopped

The person SHALL be able to stop the answer being given; the assistant SHALL
stop working on it. A request still waiting its turn SHALL be dropped.

#### Scenario: stopping

- **WHEN** somebody stops a running answer
- **THEN** the assistant stops, and the chat says it was stopped

### Requirement: An answer the model cannot give ends with what went wrong

When the model refuses or keeps failing, the answer SHALL end, show what
went wrong in the person's chat, and stop the agent, rather than wait
without end: at once for a refusal, and after the second retry otherwise.

#### Scenario: the organisation's allowance is used up

- **WHEN** the gateway refuses the agent's request
- **THEN** the chat SHALL show the reason and stop working
- **AND** the agent's session SHALL be interrupted

#### Scenario: the model keeps failing

- **WHEN** the model fails and the agent has retried twice
- **THEN** the chat SHALL show the reason and stop working
