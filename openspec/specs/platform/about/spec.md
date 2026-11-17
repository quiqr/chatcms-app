# platform/about Specification

## Purpose
Lets whoever administers an installation see what it is running and on what,
under which license, where its source lives, and who made it.

## Requirements

### Requirement: The installation says which version it runs

The about answer SHALL carry the version from `VERSION` and, when the build
knew it, the short identifier of the source it was built from, marked when
that source had uncommitted changes. A build that did not know its source
SHALL say it is a development build.

#### Scenario: A deployed build

- **WHEN** the server was built by the flake's package from a clean tree at
  a3bab43 with `VERSION` 0.1.0
- **THEN** the screen says "0.1.0 · a3bab43"

#### Scenario: A build from changed files

- **WHEN** the tree had uncommitted changes when it was built
- **THEN** the screen says the source was changed

#### Scenario: A development build

- **WHEN** the server was built without a source identifier
- **THEN** the screen says "development build" beside the version

#### Scenario: The command line agrees

- **WHEN** somebody runs `chatcmsd --version`
- **THEN** it prints the same version and identifier as the screen

### Requirement: The installation says what it runs on

The about answer SHALL carry the versions of Hugo, git, PostgreSQL and Go the
server uses, found out when it is asked. A version that cannot be found out
SHALL be given as unknown, and the rest SHALL still be given.

#### Scenario: Everything answers

- **WHEN** the about screen is opened
- **THEN** it names the Hugo, git, PostgreSQL and Go versions

#### Scenario: Hugo does not answer

- **WHEN** Hugo cannot say its version
- **THEN** Hugo is shown as unknown
- **AND** the other versions are shown

### Requirement: The license and the source are named

The about screen SHALL say the license is Apache 2.0 and SHALL let the full
text be read on the screen. It SHALL link to the GitHub page, the changelog
and the place to report a problem, each opening outside the app.

#### Scenario: Reading the license

- **WHEN** somebody opens the license on the about screen
- **THEN** the full Apache 2.0 text that ships with the build is shown

#### Scenario: The links

- **WHEN** the about screen is shown
- **THEN** it links to github.com/quiqr/chatcms-app, its changelog and its
  issues

### Requirement: The makers are credited

The about screen SHALL say ChatCMS is made by the Quiqr Team with help from
its contributors, and SHALL link to the contributors page on GitHub. The
server SHALL NOT ask GitHub for anything to show it.

#### Scenario: Credits offline

- **WHEN** the installation has no network and the about screen is opened
- **THEN** the credits are shown in full, with the link

### Requirement: Only an administrator sees it

The about screen and its answer SHALL be reached only by somebody who
administers the installation, from an entry in the platform admin area.

#### Scenario: An administrator

- **WHEN** a platform admin opens the admin area
- **THEN** About is an entry after History, and opens the screen

#### Scenario: Anybody else

- **WHEN** somebody who does not administer the installation asks for the
  about answer
- **THEN** it is refused

#### Scenario: On a phone

- **WHEN** the about screen is shown 390 px wide
- **THEN** nothing scrolls sideways and every target is at least 44 px
