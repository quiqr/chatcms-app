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

Every page the agent changed while answering SHALL be listed in the panel,
with a way to show the page and a way to undo the change. Undo SHALL go back
to the version before the change, as a new version, without publishing.

#### Scenario: a change

- **WHEN** the agent changes the home page's title
- **THEN** the panel SHALL list the home page with Show and Undo
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
