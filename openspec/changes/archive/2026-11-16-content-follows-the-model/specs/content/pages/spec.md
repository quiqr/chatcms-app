# Spec Delta

## MODIFIED Requirements

### Requirement: A file is named the way the old Quiqr names it

A page's file SHALL be named from its title: small letters, digits and hyphens,
with accents folded to the letter they are made from. In a collection under
`content/` the page SHALL be a page bundle, a folder of that name holding
`index` with the collection's extension, as the old Quiqr writes it. Elsewhere
it SHALL be a file of that name.

#### Scenario: an ordinary title

- **WHEN** a page is called "Spring prices" in a collection under `content/`
- **THEN** it SHALL be written as `spring-prices/index` with the collection's
  extension

#### Scenario: a collection outside content

- **WHEN** a page is called "Spring prices" in a collection whose folder is not
  under `content/`
- **THEN** its file SHALL be named `spring-prices` with the collection's
  extension

#### Scenario: a title with accents

- **WHEN** a title carries accented letters
- **THEN** they SHALL be folded to the letters they are made from

#### Scenario: a title that would become a path

- **WHEN** a title carries a slash or dots
- **THEN** the name SHALL carry none of them
- **AND** the file SHALL be inside the collection's own folder

#### Scenario: a title with nothing to name a file from

- **WHEN** a title has no letters or digits at all
- **THEN** it SHALL be refused with a reason

#### Scenario: the same bytes as the old Quiqr

- **WHEN** a page is made with a title
- **THEN** the file SHALL hold the same bytes the old Quiqr writes for that
  title in that collection

### Requirement: A page is made in a collection

A page SHALL be made only inside a collection the model declares, holding its
title and nothing else.

#### Scenario: making one

- **WHEN** a page is made in a collection
- **THEN** a file SHALL be written holding its title and nothing else
- **AND** a version SHALL be made and a sync asked for

#### Scenario: a name already taken

- **WHEN** a page would have the name of one that is there, as a file or as a
  page bundle
- **THEN** it SHALL be refused, saying which name
- **AND** nothing SHALL be written

#### Scenario: a group that cannot grow

- **WHEN** a page is asked for in the pages group, or in a collection that is
  not there
- **THEN** the answer SHALL be that there is nothing there

### Requirement: Renaming moves the file and the title together

Renaming SHALL move the file and change the title inside it, because to a
content creator those are one thing. A page bundle SHALL be renamed by its
folder, so that everything in it moves along.

#### Scenario: renaming

- **WHEN** a page is renamed
- **THEN** its file SHALL be written under the new name and the old one removed
- **AND** the title inside it SHALL be the new one

#### Scenario: renaming a page bundle

- **WHEN** a page bundle holding a picture is renamed
- **THEN** its folder SHALL carry the new name, with `index` and the picture
  inside it
- **AND** the old folder SHALL be gone

#### Scenario: renaming to the same name

- **WHEN** a page is renamed to what it is already called
- **THEN** nothing SHALL be written

#### Scenario: a name already taken

- **WHEN** the new name is one another page has
- **THEN** it SHALL be refused and nothing SHALL change

### Requirement: A copy leaves the first one alone

Copying SHALL make a second page and leave the first exactly as it was. A
copy of a page bundle SHALL hold everything the bundle holds.

#### Scenario: copying

- **WHEN** a page is copied
- **THEN** a second file SHALL be written under the new name
- **AND** the page it was copied from SHALL still be there

#### Scenario: copying a page bundle

- **WHEN** a page bundle holding a picture is copied
- **THEN** the new folder SHALL hold `index` and the picture

### Requirement: Removing says what will happen first

Removing SHALL be confirmed, and the confirmation SHALL say what happens to
the page and how it can be brought back. Removing a page bundle SHALL remove
its folder.

#### Scenario: what somebody is told

- **WHEN** removing is offered
- **THEN** it SHALL say that the page goes at the next publish and can be
  brought back

#### Scenario: removing

- **WHEN** a page is removed
- **THEN** the file SHALL go and a version SHALL be made
- **AND** a sync SHALL be asked for

#### Scenario: removing a page bundle

- **WHEN** a page bundle holding a picture is removed
- **THEN** its folder and everything in it SHALL go, in one version
