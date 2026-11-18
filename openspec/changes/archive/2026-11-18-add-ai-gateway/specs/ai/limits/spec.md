# Spec Delta

## Purpose

Defines how much an organisation may use the model, and what happens past it.

## ADDED Requirements

### Requirement: An organisation may be limited per minute and per month

The platform admin SHALL be able to set, per organisation, how many requests
a minute and how many tokens a calendar month it may use, or no limit. A new
organisation SHALL have no limit.

#### Scenario: setting limits

- **WHEN** the platform admin sets 30 requests a minute and 2,000,000 tokens a
  month for an organisation
- **THEN** those limits SHALL apply to every workspace of that organisation

#### Scenario: somebody else

- **WHEN** anybody but the platform admin tries to set them
- **THEN** it SHALL be refused

### Requirement: Past a limit, a request is refused with a plain reason

A request past its organisation's limit SHALL be refused before anything is
sent to the provider, saying which limit was reached in plain words.

#### Scenario: over the monthly budget

- **WHEN** an organisation has used its tokens for the month
- **THEN** its next request SHALL be refused, saying the monthly budget is
  used up

#### Scenario: too many a minute

- **WHEN** an organisation has made as many requests this minute as allowed
- **THEN** the next SHALL be refused, saying to wait a moment
