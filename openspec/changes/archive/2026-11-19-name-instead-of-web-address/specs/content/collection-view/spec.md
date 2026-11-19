## MODIFIED Requirements

### Requirement: Making an item checks its address first

The new-item dialog SHALL ask for a title, and a folder when the collection
has subfolders. On a Hugo site it SHALL show the web address made from the
title; on any other site, where that address cannot be known, it SHALL show
the name made from the title instead. Either can be changed. For those who
see technical detail it SHALL show the file that will be made. A taken
address SHALL be refused before anything is written, and the dialog SHALL
say how to fix it.

#### Scenario: Making a post

- **WHEN** somebody types "Spring sale" and chooses Create and edit
- **THEN** the item is made and opens in the editor

#### Scenario: A book that is not a Hugo site

- **WHEN** somebody makes a chapter in a Quarto book
- **THEN** the dialog shows "Name" with the name made from the title, and no
  web address

#### Scenario: A taken address

- **WHEN** the address made from the title is already used
- **THEN** the dialog says there is already a post at that address and to
  change the title or the address
- **AND** Create and edit stays unavailable

#### Scenario: A viewer

- **WHEN** a viewer lists a collection
- **THEN** no new, rename, duplicate or remove action is offered
