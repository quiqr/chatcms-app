# platform/local-development Specification

## Purpose
Defines what a development machine gets that a deployment never does: people
seeded for every role, and signing in as one of them in a single step.

## Requirements

### Requirement: Every role has a seeded person who can sign in

The development seed SHALL leave one person per role able to sign in with the
same passphrase: one who administers the installation, an implementer, an
editor and a viewer. The editor and the viewer SHALL hold their role on a
project that has a site.

#### Scenario: The editor signs in after seeding

- **WHEN** the database has been seeded and the editor's address is entered
  with the seeded passphrase
- **THEN** sign-in SHALL succeed
- **AND** the editor SHALL see the project with a site with the editor role

#### Scenario: The viewer signs in after seeding

- **WHEN** the database has been seeded and the viewer's address is entered
  with the seeded passphrase
- **THEN** sign-in SHALL succeed
- **AND** the viewer SHALL see that project with the viewer role

### Requirement: One seeded invitation stays open

The seed SHALL also invite one person who does not accept, so that an open
invitation can still be looked at on a development machine and in the
acceptance suite.

#### Scenario: The open invitation after seeding

- **WHEN** the database has been seeded
- **THEN** one invited person on the project with a site SHALL still be to
  accept
- **AND** signing in as that person SHALL be refused

### Requirement: The development server offers sign-in as each role

The sign-in screen SHALL, when served by the frontend development server, show
a block marked as development with one button per seeded role. A button SHALL
sign in as that person through the ordinary password sign-in.

#### Scenario: Signing in as the editor in one click

- **WHEN** the sign-in screen is opened from the development server and the
  editor button is pressed
- **THEN** the person SHALL be signed in as the seeded editor
- **AND** the request SHALL be the same password sign-in a typed address and
  passphrase make

#### Scenario: A seeded person who is missing

- **WHEN** a button is pressed and that person does not exist in the database
- **THEN** the screen SHALL show the same refusal as a wrong password

### Requirement: A production build carries nothing of development sign-in

A production build of the frontend SHALL contain neither the development
block nor the seeded passphrase, and the frontend check SHALL fail when it
does.

#### Scenario: The sign-in screen in a deployment

- **WHEN** the sign-in screen is served by a built binary
- **THEN** no development block SHALL be shown

#### Scenario: The passphrase leaks into the bundle

- **WHEN** the production bundle contains the seeded passphrase
- **THEN** the frontend check SHALL fail

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

### Requirement: The development screens link to the instance's inbox

When served by the frontend development server, the sign-in screen and the
forgot-password screen SHALL show a link, marked as development, to the
Mailpit inbox of the instance. The link SHALL use the host the page was opened
on and the inbox port of the instance. A production build SHALL show neither
link.

#### Scenario: The usual instance on this machine

- **WHEN** the sign-in screen is opened from the development server at
  `localhost` and the instance's inbox port is 8025
- **THEN** it SHALL link to `http://localhost:8025`

#### Scenario: A named instance opened from a phone

- **WHEN** the forgot-password screen is opened from the development server at
  `192.168.1.20` and the instance's inbox port is 8026
- **THEN** it SHALL link to `http://192.168.1.20:8026`

#### Scenario: Started without the script

- **WHEN** the frontend development server runs without an inbox port given
- **THEN** the link SHALL use port 8025

#### Scenario: In a deployment

- **WHEN** the sign-in or forgot-password screen is served by a built binary
- **THEN** no inbox link SHALL be shown

### Requirement: The frontend script passes the inbox port on

`scripts/dev-fe.sh` SHALL hand the frontend development server the inbox port
of the same instance `scripts/dev-be.sh` starts Mailpit for.

#### Scenario: A named instance

- **WHEN** `scripts/dev-fe.sh` runs with `CHATCMS_INSTANCE=test`
- **THEN** the frontend development server SHALL receive inbox port 8026
