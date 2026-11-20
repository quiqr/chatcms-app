# content/project-files Specification

## Purpose
TBD - created by archiving change project-files. Update Purpose after archive.

## Requirements

### Requirement: Implementers can browse a project's files

A project SHALL have a Files screen, reached from its Setup navigation, that
shows the directory tree of the person's own copy of the site, without
`.git` and build output. It SHALL be shown only to somebody who may see the
project's technical detail; the server SHALL refuse anybody else with 403.

#### Scenario: an implementer

- **WHEN** an implementer opens Files
- **THEN** the tree SHALL list the site's folders and files, folders first,
  each in name order
- **AND** `.git` SHALL not be listed

#### Scenario: an editor

- **WHEN** an editor asks the server for the files of a project
- **THEN** the server SHALL refuse with 403

### Requirement: The tree is usable with a keyboard

The tree SHALL follow the ARIA tree pattern: one tab stop, arrow up and down
to move, right to open a folder or go into it, left to close it or go to its
parent, and Enter to open a file.

#### Scenario: opening a file with the keyboard

- **WHEN** an implementer moves to a file with the arrow keys and presses
  Enter
- **THEN** the file SHALL open in the editor

### Requirement: A file opens in a code editor

Opening a text file SHALL show it in a code editor with line numbers,
search and highlighting by its kind; Go template actions in HTML SHALL be
marked. A binary file or one over 1 MB SHALL say it cannot be opened here.

#### Scenario: a template

- **WHEN** an implementer opens `layouts/index.html`
- **THEN** its text SHALL be shown with line numbers

#### Scenario: a picture

- **WHEN** an implementer opens a PNG
- **THEN** the screen SHALL say this file cannot be opened here

### Requirement: Saving a file is a change like any other

Saving SHALL write the file into the person's own copy as a saved change
that is published like every other change. A file that changed since it was
opened SHALL NOT be overwritten; the person SHALL be told and offered to
open it again.

#### Scenario: saving a template

- **WHEN** an implementer changes a template and saves
- **THEN** the change SHALL be counted as waiting to be published
- **AND** publishing SHALL put the new template on the remote

#### Scenario: a file that changed meanwhile

- **WHEN** the file changed in the person's copy after it was opened
- **THEN** saving SHALL be refused with a message saying so

### Requirement: A path cannot leave the project

The server SHALL refuse a path that is absolute, contains `..`, passes
through a symbolic link, or lies in `.git`, for reading and for writing.

#### Scenario: a path out of the project

- **WHEN** a file is asked for as `../other/config.toml`
- **THEN** the server SHALL refuse it

#### Scenario: a symbolic link

- **WHEN** a path passes through a symbolic link
- **THEN** the server SHALL refuse it
