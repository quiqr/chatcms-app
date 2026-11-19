# preview/origin Specification

## Purpose
Defines where previews are served from and how a preview frame is let in, so
that a site's own JavaScript never runs where the app's session lives.

## Requirements

### Requirement: Previews may be served from an origin of their own

When an installation names a preview address, previews SHALL be served from
that origin only, the app's origin SHALL NOT serve them, and the preview
origin SHALL serve nothing but previews.

#### Scenario: the app's origin

- **WHEN** a preview is asked for on the app's own address
- **THEN** there SHALL be nothing there

#### Scenario: the preview origin

- **WHEN** anything but a preview is asked for on the preview address
- **THEN** there SHALL be nothing there

#### Scenario: no preview address

- **WHEN** an installation names no preview address
- **THEN** previews SHALL be served from the app's origin as before

### Requirement: A preview on its own origin is reached by an unguessable address

On the preview origin, a preview SHALL be reached only by an address holding
a key the server made at random when the workspace started, which ends when
the workspace sleeps. Every request SHALL be checked against the right of the
workspace's person to read the project. The preview origin SHALL use no
cookie and never see the app's session.

#### Scenario: the workspace's address

- **WHEN** the frame opens the preview at the address the app gave it
- **THEN** the preview SHALL be served, links and live reload included

#### Scenario: a wrong or old key

- **WHEN** an address holds a key no warm workspace has
- **THEN** there SHALL be nothing there

#### Scenario: the right to read

- **WHEN** the workspace's person may no longer read the project
- **THEN** the preview SHALL be refused

### Requirement: Without an origin of its own, a preview follows the app's address

Without a preview origin, a preview SHALL be reached through the address the
browser reached the app by, whatever name that is, so the person's session
reaches it.

#### Scenario: the app reached by another name

- **WHEN** the app's base URL is on `127.0.0.1` and a person uses it through
  `localhost`
- **THEN** their preview SHALL show, and a selection in it SHALL reach the
  app
