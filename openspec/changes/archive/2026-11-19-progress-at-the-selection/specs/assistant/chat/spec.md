## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: A running answer can be stopped

The person SHALL be able to stop the answer being given; the assistant SHALL
stop working on it. A request still waiting its turn SHALL be dropped.

#### Scenario: stopping

- **WHEN** somebody stops a running answer
- **THEN** the assistant stops, and the chat says it was stopped
