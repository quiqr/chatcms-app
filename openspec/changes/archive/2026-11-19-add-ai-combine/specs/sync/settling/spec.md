# Spec Delta

## ADDED Requirements

### Requirement: The AI may combine two versions of a field

Where the assistant is available, a field two people changed SHALL offer to
let the AI combine them. The combined text SHALL be shown to the person, who
MAY change it and use it, or still keep either version. A combined text
SHALL be written only when the answers are saved. The request SHALL count
against the organisation's AI limits like any other.

#### Scenario: combining two titles

- **WHEN** two people changed a page's title and one asks the AI to combine
  them
- **THEN** a combined title SHALL be shown, and saving the answers with it
  SHALL keep that title

#### Scenario: no assistant

- **WHEN** the installation has no assistant
- **THEN** combining SHALL NOT be offered

#### Scenario: the limit is reached

- **WHEN** the organisation has used its AI allowance
- **THEN** combining SHALL say so, and the two versions SHALL still be
  offered
