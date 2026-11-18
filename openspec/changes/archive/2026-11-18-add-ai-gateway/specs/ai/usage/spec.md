# Spec Delta

## Purpose

Defines what is recorded about each call to the model and who may see it.

## ADDED Requirements

### Requirement: Every call is recorded

Every request passed to the provider SHALL be recorded with its organisation,
project, person, model, tokens in and out, and how it ended.

#### Scenario: a call

- **WHEN** a request is answered
- **THEN** a record SHALL be kept with the tokens the provider reported

#### Scenario: a streamed call

- **WHEN** a streamed request is answered
- **THEN** its record SHALL carry the tokens of the whole answer

### Requirement: The platform admin sees use per organisation

The platform admin SHALL see, per organisation, the requests and tokens of
the current month beside its limits. Nobody else SHALL be given them.

#### Scenario: the platform admin

- **WHEN** the platform admin opens an organisation
- **THEN** this month's requests and tokens SHALL be shown beside the limits
