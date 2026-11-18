# workspace/hugo-version Specification

## Purpose
Defines how a project's own Hugo version is found, made available and used.

## Requirements

### Requirement: A project's preview and build use its own Hugo

A project that records a Hugo version SHALL be previewed and built by that
version of Hugo. A project that records none SHALL use the installation's
Hugo.

#### Scenario: a recorded version

- **WHEN** a project records Hugo 0.139.0 and its preview or a build runs
- **THEN** it SHALL run Hugo 0.139.0

#### Scenario: no recorded version

- **WHEN** a project records no Hugo version
- **THEN** the installation's Hugo SHALL be used

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
