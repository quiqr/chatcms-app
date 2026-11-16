# Spec Delta

## MODIFIED Requirements

### Requirement: A site arrives from a git repository, an export, or a template

Importing SHALL accept a git repository address, a directory holding a site
exported from the Quiqr desktop app, or a template from the template list.

#### Scenario: From a repository

- **WHEN** a repository address is given
- **THEN** the site is fetched into the project

#### Scenario: From an export

- **WHEN** a directory holding an exported site is given
- **THEN** the site becomes the project's content
- **AND** it is sent to the project's remote

#### Scenario: From a template

- **WHEN** a template is chosen and the project's remote is an empty
  repository the deploy key can reach
- **THEN** the site starts as that template's files in one version named
  after the template, with none of the template's history
- **AND** it is sent to the project's remote

#### Scenario: From a template into a repository that is not empty

- **WHEN** a template is chosen and the project's remote already holds work
- **THEN** importing SHALL be refused with a message saying the repository
  is not empty
- **AND** nothing SHALL be sent to it

#### Scenario: From a template that is not in the list

- **WHEN** the chosen template is not in the template list
- **THEN** importing SHALL be refused

#### Scenario: An export and its repository agree

- **WHEN** the same site is imported once as an export and once as the
  repository made from that export
- **THEN** the two projects hold the same content
