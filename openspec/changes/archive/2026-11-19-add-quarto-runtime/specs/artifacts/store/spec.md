# Spec Delta

## Purpose

Defines how artifacts are kept, listed and downloaded, and where they are
shown.

## ADDED Requirements

### Requirement: Artifacts are kept per publish

Each artifact a publish produces SHALL be kept with its project, the publish
it came from, its name, size and time. A later publish SHALL keep a new
artifact and leave the earlier one. A file over 200 MB SHALL NOT be kept,
and the publish SHALL say so.

#### Scenario: two publishes

- **WHEN** a project is published twice and each publish produces
  `book.pdf`
- **THEN** both SHALL be kept, newest first

### Requirement: Anybody who may read the project may download its artifacts

Artifacts SHALL be listed and downloaded by anybody who may read the
project's content, and by nobody else. A download SHALL carry the
artifact's name.

#### Scenario: a viewer downloads

- **WHEN** a viewer of the project downloads an artifact
- **THEN** they SHALL get the file under its name

#### Scenario: somebody from another organisation

- **WHEN** a person who may not see the project asks for an artifact
- **THEN** it SHALL be as if the artifact does not exist

### Requirement: Artifacts show where a person looks for them

The Site navigation SHALL show Artifacts after History once the project has
an artifact. The publish Done state SHALL offer the downloads of that
publish.

#### Scenario: a project without artifacts

- **WHEN** a project has never kept an artifact
- **THEN** the navigation SHALL show no Artifacts item

#### Scenario: a publish with a PDF

- **WHEN** a publish kept `book.pdf`
- **THEN** its Done state SHALL offer `book.pdf` to download
