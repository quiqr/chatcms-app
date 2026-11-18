# Spec Delta

## MODIFIED Requirements

### Requirement: Asking twice uses the preview already running

One person SHALL have at most one preview of a site running.

#### Scenario: a second request

- **WHEN** a person asks for a preview of the same site again
- **THEN** the one already running for them SHALL be used

#### Scenario: another person

- **WHEN** another person asks for a preview of the same site
- **THEN** a preview of that person's own copy SHALL be used

### Requirement: A preview nobody is looking at is stopped

A preview SHALL be stopped when nobody has asked for it for a stated length of
time, by its workspace going to sleep.

#### Scenario: nobody looking

- **WHEN** nobody has asked for a preview for longer than that
- **THEN** it SHALL be stopped and its resources released

#### Scenario: somebody looking

- **WHEN** somebody asks for it again
- **THEN** it SHALL be kept, and SHALL still answer

### Requirement: Only so many previews run at once

The number of previews running at once SHALL be capped by the number of
workspaces allowed to be warm at once, so that one busy afternoon cannot
exhaust the machine.

#### Scenario: the cap

- **WHEN** as many workspaces are starting or warm as are allowed
- **THEN** another preview SHALL be refused with a reason worth showing
- **AND** the ones running SHALL be untouched

### Requirement: Nothing outlives the app

When the app stops, every preview whose workspace runs as part of the app
SHALL be stopped, and stopping SHALL be safe to ask for more than once.

#### Scenario: the app stops

- **WHEN** the app stops while previews run on this machine as part of it
- **THEN** every one of them SHALL be stopped

#### Scenario: stopping twice

- **WHEN** stopping is asked for again
- **THEN** it SHALL do nothing rather than fail
