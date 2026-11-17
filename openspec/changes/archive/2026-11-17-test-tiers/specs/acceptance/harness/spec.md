## ADDED Requirements

### Requirement: The suite runs its tests side by side

The acceptance suite SHALL run several tests at once, each still owning its
own PostgreSQL, remote, site, Mailpit and server. The number at once SHALL
default to half the available cores, at least one, and SHALL be settable with
`E2E_WORKERS`. A failed test SHALL NOT be retried.

#### Scenario: Several at once

- **WHEN** the suite runs on a machine with twelve cores
- **THEN** six tests run at the same time
- **AND** every test passes as it does alone

#### Scenario: Set by hand

- **WHEN** the suite runs with `E2E_WORKERS=1`
- **THEN** the tests run one after another

### Requirement: A run's ports are free ones

Every port a run listens on SHALL be one the system has handed out as free,
not one chosen at random.

#### Scenario: Two runs at once

- **WHEN** two tests start their servers and Mailpits at the same moment
- **THEN** no two of them listen on the same port
