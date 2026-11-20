# history/project-log Specification

## Purpose
TBD - created by archiving change project-log. Update Purpose after archive.

## Requirements

### Requirement: Every preview start is kept

Each start of a person's workspace SHALL be kept with when it began, who
started it, the backend, whether it succeeded and how long it took, its
steps, its log (at most its last 400 lines), and what Hugo it used: the
version the project asks for, the program chosen, why that program was
chosen, and the first line the program printed when it ran. Starts SHALL be
removed after the configured history days.

#### Scenario: a pinned version

- **WHEN** a project asks for Hugo 0.101.0 and the installation pins versions
- **THEN** the kept start SHALL name 0.101.0 as asked, the pinned program as
  chosen, pinning as the reason, and the version the program printed

#### Scenario: pinning is off

- **WHEN** a project asks for Hugo 0.101.0 and the installation does not pin
- **THEN** the kept start SHALL name the installed program and say that
  pinning is off

#### Scenario: a failed start

- **WHEN** a start fails
- **THEN** it SHALL be kept with its reason and its log

### Requirement: A project has a Log page for implementers

A project SHALL have a Log page, reached from its Setup navigation, listing
its preview starts, builds and publications newest first, each with its kind,
time, who, outcome and duration, and filterable by kind. It SHALL be shown
only to somebody who may see the project's technical detail; the server SHALL
refuse anybody else with 403.

#### Scenario: an implementer reads the log

- **WHEN** an implementer opens Log
- **THEN** preview starts, builds and publications SHALL be listed together,
  newest first
- **AND** choosing "Preview starts" SHALL list only those

#### Scenario: an editor

- **WHEN** an editor looks at the project's navigation
- **THEN** there SHALL be no Log
- **AND** asking the server for the log SHALL be refused with 403

### Requirement: An entry opens to its detail and has its own address

Opening an entry SHALL show its detail: for a start its Hugo facts, steps and
log; for a build its log and exit; for a publication its steps and what went
wrong. Each entry SHALL have an address that opens it directly, at 1440 px
and at 390 px, and a log SHALL be copyable.

#### Scenario: linked from a failed start

- **WHEN** an implementer follows "See what went wrong" from the preview
- **THEN** the Log page SHALL open with that start unfolded

#### Scenario: on a phone

- **WHEN** an entry is opened at 390 px
- **THEN** its detail SHALL fill the screen with a way back to the list

#### Scenario: a start still running

- **WHEN** an implementer opens a start that is still running
- **THEN** its steps and log SHALL follow as it goes
