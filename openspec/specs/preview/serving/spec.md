# preview/serving Specification

## Purpose
Defines how a preview is reached and who may see one.

## Requirements

### Requirement: A preview is reached through the app

A preview SHALL be served under the app's own address, and the preview server
itself SHALL listen only on the machine it runs on.

#### Scenario: somebody not signed in

- **WHEN** a preview is asked for by somebody not signed in
- **THEN** it SHALL be refused
- **AND** no part of the site SHALL be served

#### Scenario: the path within the site

- **WHEN** a preview is asked for a path inside the site
- **THEN** the preview SHALL be asked for that path, without the part of the
  address that names the project

#### Scenario: the front page

- **WHEN** a preview is asked for with no path at all
- **THEN** the site's front page SHALL be served

#### Scenario: an address naming no project

- **WHEN** the address names no project
- **THEN** the answer SHALL be that there is nothing there

### Requirement: Anybody who may read the project may preview it

Seeing a preview SHALL need the right to read the project and nothing more.

#### Scenario: a viewer

- **WHEN** somebody who may read but not change the site asks for a preview
- **THEN** they SHALL be shown it

#### Scenario: somebody with no part in the project

- **WHEN** somebody with no part in the project asks
- **THEN** they SHALL be told there is nothing there

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

### Requirement: A preview shows the asking person's own copy

A preview SHALL show the site as it is in the asking person's own copy, their
saves that have not yet reached the shared version included.

#### Scenario: two people, two changes

- **WHEN** two people each save a different change to the same page and both
  look at its preview
- **THEN** each SHALL see their own change in their preview
- **AND** neither SHALL see the other's change until it has reached their
  copy

### Requirement: A theme's own absolute addresses work under any name

The preview SHALL turn the absolute address it gave the site generator as
its base URL into a path, in the pages, styles, scripts and feeds it serves,
so a theme that writes its addresses with `{{ .Site.BaseURL }}` shows its
styles and pictures however the person reached the app. Other addresses and
binary files SHALL be left as they are.

#### Scenario: the app opened under another name

- **WHEN** the app's base URL is `http://127.0.0.1:8080` and a person opens it
  as `http://localhost:5173`
- **AND** a theme links `{{ .Site.BaseURL }}css/styles.css`
- **THEN** the preview SHALL link `/preview/<organisation>/<project>/css/styles.css`
  and the style SHALL apply
