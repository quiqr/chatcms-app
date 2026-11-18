# Spec Delta

## MODIFIED Requirements

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
