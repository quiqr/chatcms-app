# Spec Delta

## Purpose

Defines how the assistant's acceptance tests run against a real model.

## ADDED Requirements

### Requirement: The assistant's tests can run against a real model

With `E2E_REAL_AI` set, the assistant's acceptance tests SHALL use the model
named in the developer's AI settings file instead of the scripted one, and
SHALL check only that the page asked about changed as asked, shows as
changed, and goes back on Undo. Each test SHALL say how many calls and
tokens it used. Without it they SHALL use the scripted model.

#### Scenario: a real run

- **WHEN** a developer runs the assistant specs with `E2E_REAL_AI=1`
- **THEN** they SHALL run against the model in `~/.config/chatcms/ai.env`,
  and say what each test used

#### Scenario: the gate

- **WHEN** the gate runs the acceptance tests
- **THEN** they SHALL use the scripted model
