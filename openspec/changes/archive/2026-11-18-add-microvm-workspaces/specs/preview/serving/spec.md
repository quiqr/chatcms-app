# Spec Delta

## MODIFIED Requirements

### Requirement: A preview shows the site as a reader will see it

A preview SHALL serve the site built by the same Hugo that builds it for
publishing: the project's own version when it records one, the
installation's otherwise.

#### Scenario: the site renders

- **WHEN** a preview of a site that builds is asked for
- **THEN** the site's own pages SHALL be served

#### Scenario: a site that will not build

- **WHEN** a preview of a site Hugo cannot build is asked for
- **THEN** it SHALL be refused with a reason worth showing
- **AND** no preview SHALL be left running

#### Scenario: a project with its own Hugo

- **WHEN** a project records a Hugo version and its preview is asked for
- **THEN** the preview SHALL be built by that version
