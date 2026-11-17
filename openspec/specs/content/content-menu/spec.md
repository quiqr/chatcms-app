# content/content-menu Specification

## Purpose
TBD - created by archiving change content-menu-screens. Update Purpose after archive.

## Requirements

### Requirement: The content column shows the menu

Inside a project the content area SHALL have a column holding the content
menu, as content/content-navigation derives it: a heading with how many
pages and collections there are, then each group in model order.

#### Scenario: Groups and rows

- **WHEN** a site's menu has two groups, one with two singles and one with a
  collection of 48 items
- **THEN** the column shows both groups in that order
- **AND** each single is a row with a page icon, and the collection a row with
  a stack icon and 48

#### Scenario: Waiting to be published

- **WHEN** a single, or any item of a collection, has a change nobody has
  published
- **THEN** its row carries a dot saying so

#### Scenario: The open item's collection

- **WHEN** an item of a collection is open
- **THEN** that collection's row is shown as the current one

#### Scenario: A long title

- **WHEN** a row's title does not fit
- **THEN** it is cut short with an ellipsis and the whole title is offered on
  hover

### Requirement: A group folds and stays folded

Each group SHALL fold and unfold from its heading. A folded group SHALL show
how many rows it holds. The choice SHALL be remembered for that person and
project in that browser.

#### Scenario: Folding a group

- **WHEN** somebody folds a group and reloads the screen
- **THEN** the group is still folded and shows its count

### Requirement: The content column folds by hand

On screens 1024 px and wider, the content column SHALL have an icon in its
heading that folds it, on the collection view and in the editor alike. The
folded column SHALL be a narrow strip in the column's place, holding an icon
that unfolds it. The state SHALL be remembered per person, across projects and
reloads, and SHALL change only when the person folds or unfolds it.

#### Scenario: Folding and reloading

- **WHEN** somebody folds the column in the editor and reloads
- **THEN** the editor opens with the column folded to its strip
- **AND** the strip's unfold icon brings the column back

#### Scenario: Folding on the collection view

- **WHEN** somebody folds the column on the collection view and opens an item
- **THEN** the editor shows the column folded too

#### Scenario: An earlier choice is kept

- **WHEN** somebody folded the column in the editor before folding became
  available on every screen
- **THEN** the column is folded for them on every screen

### Requirement: Quick find opens any page by its title

⌘K or Ctrl+K anywhere in a project SHALL open a search over the titles of all
singles and all collection items the person may read. Results SHALL be
grouped by group and collection, and Enter SHALL open the first.

#### Scenario: Finding an item

- **WHEN** somebody presses Ctrl+K and types "spring"
- **THEN** the items whose titles hold "spring" are offered, under their
  collection
- **AND** Enter opens the first

#### Scenario: Nothing found

- **WHEN** nothing matches
- **THEN** quick find says so in plain words

#### Scenario: A developer's section

- **WHEN** an editor searches for a single that only a developer's section
  names
- **THEN** it is not offered

### Requirement: Narrow screens reach the same content

Below 1024 px the content column SHALL be a drawer, opened from the icon in
the column's strip and closing when something is chosen. Below 640 px content
SHALL be three screens (the menu, a collection, the editor) inside the phone
frame and its existing tab bar.

#### Scenario: A tablet

- **WHEN** the screen is 800 px wide and somebody opens the drawer from the
  strip and picks a collection
- **THEN** the drawer closes and the collection is shown

#### Scenario: A phone

- **WHEN** the screen is 390 px wide
- **THEN** Content shows the grouped menu, a collection opens as a screen of
  its own, and an item opens in the phone editor
- **AND** nothing scrolls sideways and every target is at least 44 px

### Requirement: The menu needs only the right to read

Seeing the content column and quick find SHALL need only the right to read the
project.

#### Scenario: A viewer

- **WHEN** a viewer opens a project
- **THEN** the content column and quick find work for them
