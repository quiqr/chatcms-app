## MODIFIED Requirements

### Requirement: Implementers can see the steps and the log

The steps of a start (starting the workspace, fetching the site, building the
preview) and the state of each SHALL be given to somebody who may see a
project's technical detail. The log of a start SHALL be read on the project's
Log page, which the preview column links to for them. Nobody else SHALL be
given the steps, the log or the link.

#### Scenario: an implementer

- **WHEN** an implementer looks at the preparing indicator
- **THEN** it SHALL offer "Show the log", which opens that start on the
  project's Log page

#### Scenario: a start that failed

- **WHEN** an implementer sees that the preview did not start
- **THEN** it SHALL offer "See what went wrong", which opens that start on
  the Log page

#### Scenario: a content creator

- **WHEN** a content creator looks at the preparing indicator
- **THEN** there SHALL be no link to the log
- **AND** what the server answers about the start SHALL leave the steps and
  the log out
