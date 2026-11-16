# Spec Delta

## ADDED Requirements

### Requirement: The module sets the template feed

The module SHALL take the address of the template feed, defaulting to the
Quiqr community feed, and SHALL allow it to be switched off.

#### Scenario: Left at its default

- **WHEN** a machine sets no template feed
- **THEN** the gallery SHALL read the Quiqr community feed

#### Scenario: Switched off

- **WHEN** a machine sets the template feed to nothing
- **THEN** no feed SHALL be fetched
- **AND** creating a project SHALL not offer starting from a template
