# Spec Delta

## ADDED Requirements

### Requirement: The backend and the frontend each start with their own script

A development machine SHALL start the backend with `scripts/dev-be.sh` and the
frontend development server with `scripts/dev-fe.sh`. Each script SHALL list
its options at its top, and both SHALL refuse to run outside the development
shell.

#### Scenario: Starting the frontend on a fresh checkout

- **WHEN** `scripts/dev-fe.sh` runs inside the development shell and
  `web/node_modules` does not exist
- **THEN** the frontend dependencies SHALL be installed from the lock file
- **AND** the frontend development server SHALL start

#### Scenario: Starting the frontend outside the development shell

- **WHEN** `scripts/dev-fe.sh` runs where the development shell's tools are
  missing
- **THEN** it SHALL exit with a failure
- **AND** it SHALL say to run it inside `nix develop`

### Requirement: The frontend follows the backend's instance

Given the same `CHATCMS_INSTANCE`, `scripts/dev-fe.sh` SHALL send API requests
to the port `scripts/dev-be.sh` serves on, and SHALL serve on a port of its
own that another instance does not use. `CHATCMS_PORT` and `CHATCMS_WEB_PORT`
SHALL override the two ports.

#### Scenario: The usual instance

- **WHEN** `scripts/dev-fe.sh` runs without `CHATCMS_INSTANCE`
- **THEN** it SHALL send API requests to port 8080
- **AND** it SHALL serve on port 5173

#### Scenario: A named instance

- **WHEN** `scripts/dev-fe.sh` runs with `CHATCMS_INSTANCE=test`
- **THEN** it SHALL send API requests to port 8081, where
  `CHATCMS_INSTANCE=test scripts/dev-be.sh` serves
- **AND** it SHALL serve on port 5174

#### Scenario: An instance name that cannot name a database

- **WHEN** `CHATCMS_INSTANCE` holds anything but small letters and digits
- **THEN** both scripts SHALL refuse it with the same message

### Requirement: The frontend listens where the backend listens

`scripts/dev-fe.sh` SHALL listen on the address in `CHATCMS_HOST` when it is
set, so that one setting makes both halves reachable from another machine.

#### Scenario: Reachable from a phone

- **WHEN** `scripts/dev-fe.sh` runs with `CHATCMS_HOST=0.0.0.0`
- **THEN** the frontend development server SHALL listen on all addresses

### Requirement: A missing backend is named, not fatal

Before it starts, `scripts/dev-fe.sh` SHALL check whether anything answers on
the API port. When nothing does, it SHALL print a warning with the command
that starts the matching backend, and SHALL start the frontend anyway.

#### Scenario: No backend running

- **WHEN** `scripts/dev-fe.sh` runs with `CHATCMS_INSTANCE=test` and nothing
  listens on port 8081
- **THEN** it SHALL print a warning naming
  `CHATCMS_INSTANCE=test scripts/dev-be.sh`
- **AND** the frontend development server SHALL still start

#### Scenario: Backend running

- **WHEN** something answers on the API port
- **THEN** no warning SHALL be printed
