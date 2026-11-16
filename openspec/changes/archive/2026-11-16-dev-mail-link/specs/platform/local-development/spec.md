# Spec Delta

## ADDED Requirements

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
