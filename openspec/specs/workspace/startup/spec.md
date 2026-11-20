# workspace/startup Specification

## Purpose
Defines what a person sees while their workspace starts, and who may see the
steps and the log behind it.

## Requirements

### Requirement: A starting workspace says so

While a person's workspace is starting, the preview column SHALL show that
the preview is being prepared, and SHALL show the preview by itself once the
workspace is warm.

#### Scenario: preparing

- **WHEN** a person opens a project whose workspace is starting
- **THEN** the preview column SHALL say the preview is being prepared

#### Scenario: ready

- **WHEN** the workspace becomes warm
- **THEN** the preview SHALL be shown without the person doing anything

#### Scenario: did not start

- **WHEN** the workspace fails to start
- **THEN** the preview column SHALL say so with a reason worth showing

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

### Requirement: The indicator speaks without git words

What a content creator sees while a workspace starts SHALL use none of the
words the copy rules keep from them.

#### Scenario: the words

- **WHEN** the preparing indicator is shown to a content creator
- **THEN** it SHALL contain none of the forbidden words
