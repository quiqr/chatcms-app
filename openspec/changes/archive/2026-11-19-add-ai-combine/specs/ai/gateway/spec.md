# Spec Delta

## ADDED Requirements

### Requirement: The host may ask the model once

ChatCMS itself MAY ask the model a single question on a person's behalf,
outside any workspace. Such a request SHALL use the configured model, key
and headers, SHALL be refused past the organisation's limits, and SHALL be
recorded as that person's use.

#### Scenario: a combination

- **WHEN** a person asks the AI to combine two versions of a field
- **THEN** the model SHALL be asked once, and the use SHALL be recorded for
  that person's organisation
