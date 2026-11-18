# workspace/lifecycle Specification

## Purpose
Defines what a workspace is, who it belongs to, when it starts, sleeps and is
cleaned up, how many run at once, and what a restart of the server does to
them.

## Requirements

### Requirement: A workspace belongs to one person and one project

A workspace SHALL belong to exactly one person and one project. Asking for the
workspace of the same person and project again SHALL give the same workspace.

#### Scenario: two people, one project

- **WHEN** two people open the same project
- **THEN** each SHALL get a workspace of their own

#### Scenario: asking twice

- **WHEN** a person asks for their workspace of a project a second time
- **THEN** the workspace they already have SHALL be used

### Requirement: Where a workspace runs is hidden from the rest of the app

A workspace SHALL be reached by an address the workspace component hands out.
Nothing outside that component SHALL depend on which backend runs it, so that
choosing another backend changes no behaviour described here.

#### Scenario: another backend

- **WHEN** the installation runs workspaces on a different backend
- **THEN** every requirement in this capability SHALL still hold

### Requirement: A workspace is starting, warm or sleeping

A workspace SHALL be in one of three states: starting, warm or sleeping. The
state SHALL be recorded in the database, so that it survives a restart of the
server.

#### Scenario: started

- **WHEN** a workspace has started and answers
- **THEN** it SHALL be warm

#### Scenario: failed to start

- **WHEN** a workspace does not answer within the time allowed
- **THEN** it SHALL be sleeping, with the reason recorded
- **AND** nothing it started SHALL be left running

### Requirement: Opening a project starts its workspace

Opening a project SHALL start the person's workspace if it is not warm, before
the preview is asked for.

#### Scenario: opening the editor

- **WHEN** a person opens a project in the editor
- **THEN** their workspace SHALL be starting or warm before they ask for the
  preview

### Requirement: A workspace nobody uses goes to sleep

A warm workspace SHALL go to sleep when nobody has used it for a stated length
of time. Sleeping SHALL stop what runs in it and keep the person's copy of the
site, unsaved and unshared work included.

#### Scenario: nobody using it

- **WHEN** a workspace has not been used for longer than the idle time
- **THEN** it SHALL be sleeping and its processes stopped
- **AND** the person's copy of the site SHALL be kept

#### Scenario: woken again

- **WHEN** a sleeping workspace is asked for
- **THEN** it SHALL start again on the same copy of the site

### Requirement: Only so many workspaces are warm at once

The number of workspaces starting or warm at once SHALL be capped per
installation.

#### Scenario: the cap

- **WHEN** as many workspaces are starting or warm as are allowed
- **THEN** another SHALL be refused with a reason worth showing
- **AND** the ones running SHALL be untouched

### Requirement: A restart of the server leaves no workspace unaccounted for

On start-up the server SHALL go through every workspace recorded as starting
or warm. One that still runs and answers SHALL be used again; any other SHALL
be stopped and recorded as sleeping.

#### Scenario: still running

- **WHEN** the server starts and a workspace recorded as warm still answers
- **THEN** it SHALL be warm and served without being started again

#### Scenario: gone

- **WHEN** the server starts and a workspace recorded as warm no longer
  answers
- **THEN** whatever is left of it SHALL be stopped
- **AND** it SHALL be recorded as sleeping

### Requirement: A workspace nobody can use any more is cleaned up

A sleeping workspace SHALL be removed, its copy of the site with it, when its
person may no longer read the project, when the project is archived or
removed, or when it has not been used for 30 days and holds nothing that has
not reached the shared version.

#### Scenario: a person removed from the project

- **WHEN** the person of a sleeping workspace may no longer read the project
- **THEN** the workspace and its copy of the site SHALL be removed

#### Scenario: unused, with unshared work

- **WHEN** a workspace has not been used for 30 days and its copy holds work
  that has not reached the shared version
- **THEN** it SHALL be kept

### Requirement: A restart asks nothing of the people working

When the app stops, a workspace whose backend runs it apart from the app SHALL
be left running, and one whose backend runs it as part of the app SHALL be
stopped and recorded as sleeping. After a restart, a workspace SHALL be served
again without the person doing anything: as it is when it still runs, started
anew when it does not. Stopping a workspace SHALL be safe to ask for more than
once.

#### Scenario: a backend apart from the app

- **WHEN** the app stops while such a workspace is warm, and starts again
- **THEN** the workspace SHALL be served again without being started anew

#### Scenario: a backend that is part of the app

- **WHEN** the app stops while such a workspace is warm
- **THEN** it SHALL be stopped and recorded as sleeping
- **AND** asking for it after the restart SHALL start it again

#### Scenario: stopping twice

- **WHEN** stopping a workspace is asked for again
- **THEN** it SHALL do nothing rather than fail
