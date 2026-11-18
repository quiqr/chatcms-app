# workspace/hugo-version Specification

## Purpose
Defines how a project's own Hugo version is found, made available and used.

## Requirements

### Requirement: A project's preview and build use its own Hugo

A project SHALL be previewed and built by the Hugo version its runtime
definition names; failing that, by the version the project records; failing
that, by the installation's Hugo.

#### Scenario: a recorded version

- **WHEN** a project records Hugo 0.139.0, has no runtime definition, and its
  preview or a build runs
- **THEN** it SHALL run Hugo 0.139.0

#### Scenario: no recorded version

- **WHEN** a project records no Hugo version and has no runtime definition
- **THEN** the installation's Hugo SHALL be used

#### Scenario: a version in the runtime definition

- **WHEN** the runtime definition names Hugo 0.145.0 and the project records
  0.139.0
- **THEN** Hugo 0.145.0 SHALL be used

### Requirement: A version is found once and remembered

The package a version resolves to SHALL be remembered, so that a later
preview or build of the same version does not ask for it again.

#### Scenario: the second time

- **WHEN** a version was found before
- **THEN** it SHALL be used without asking outside the installation

### Requirement: A version that cannot be had is said, not guessed

When a recorded version cannot be found or made available, the preview or
build SHALL be refused with a reason worth showing, naming the version. It
SHALL NOT fall back to another version silently.

#### Scenario: an unknown version

- **WHEN** a project records a Hugo version that does not exist
- **THEN** its preview SHALL be refused, saying that version could not be
  found
