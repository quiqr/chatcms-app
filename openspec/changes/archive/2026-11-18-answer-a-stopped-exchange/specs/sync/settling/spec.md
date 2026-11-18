# Spec Delta

## ADDED Requirements

### Requirement: An open question can be asked again

While an exchange is stopped on a question, bringing changes in and asking
how the site stands SHALL put the same question again, without fetching or
replaying anything, and answering it SHALL finish the exchange.

#### Scenario: coming back to the question

- **WHEN** an exchange has stopped on a question and the person opens the
  page that asks it later
- **THEN** the same question SHALL be put, field by field

#### Scenario: answering later

- **WHEN** the person answers the question in a request after the one that
  stopped the exchange
- **THEN** the answers SHALL be applied and the exchange finished
- **AND** the person's work SHALL reach the shared version

#### Scenario: checking for updates while asked

- **WHEN** the person checks for updates while a question is open
- **THEN** they SHALL be shown the question rather than an error
