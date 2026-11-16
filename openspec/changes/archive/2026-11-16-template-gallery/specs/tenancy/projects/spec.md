# Spec Delta

## ADDED Requirements

### Requirement: A project records the Hugo version it asks for

A project SHALL record the Hugo version its site asks for. Starting from a
template SHALL set it to the template's version. Project settings SHALL show
it and allow changing it, and SHALL say when it differs from the Hugo this
installation builds with.

#### Scenario: Started from a template

- **WHEN** a project starts from a template that asks for Hugo `0.101.0`
  extended
- **THEN** the project SHALL record `0.101.0` extended

#### Scenario: Changing it

- **WHEN** somebody allowed to change settings sets another Hugo version
- **THEN** the project SHALL record that version

#### Scenario: A version this installation does not run

- **WHEN** the recorded version differs from the installation's Hugo
- **THEN** the settings SHALL say which version builds the site here
