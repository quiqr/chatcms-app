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
folded column SHALL be a 48 px strip in the column's place that is one button
unfolding it. The strip SHALL read, from top to bottom, `CONTENT`, then the
current page or collection with a collection's item count, or only `CONTENT`
when nothing is chosen, and SHALL show a dot when anything in the menu waits
to be published. A title too long for the strip SHALL be cut short and shown
whole in a tooltip. The state SHALL be remembered per person, across projects
and reloads, and SHALL change only when the person folds or unfolds it.

#### Scenario: Folding and reloading

- **WHEN** somebody folds the column in the editor and reloads
- **THEN** the editor opens with the column folded to its strip
- **AND** clicking anywhere on the strip brings the column back

#### Scenario: Folding on the collection view

- **WHEN** somebody folds the column on the collection view and opens an item
- **THEN** the editor shows the column folded too

#### Scenario: An earlier choice is kept

- **WHEN** somebody folded the column in the editor before folding became
  available on every screen
- **THEN** the column is folded for them on every screen

#### Scenario: The strip says where you are

- **WHEN** the column is folded on a collection of ten case studies
- **THEN** the strip reads `CONTENT · Case studies 10`
- **AND** on the content home it reads only `CONTENT`

#### Scenario: The strip shows waiting work

- **WHEN** a page in the menu has a change waiting to be published
- **THEN** the strip shows the dot, exactly when the unfolded menu shows one

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

### Requirement: A location picker reaches any page without the column

On screens wider than a phone, the collection view and the editor SHALL show a
breadcrumb `Content / <group> / <current>`, whose last part opens a picker,
whether the column is folded or not. Choosing an entry in the picker SHALL
open it and close the picker. While the column is hidden, folded to its strip
or shown as a drawer on a narrower screen, the picker SHALL offer "Show the
full content menu", which unfolds the column. While the column is open, the
picker SHALL NOT offer it.

#### Scenario: Switching while folded

- **WHEN** the column is folded and somebody opens the picker and chooses
  News
- **THEN** News opens and the picker closes, and the column stays folded

#### Scenario: The column is already open

- **WHEN** the column is open and somebody opens the picker
- **THEN** the picker does not offer "Show the full content menu"

### Requirement: The picker holds quick find and the compact menu

The picker SHALL hold a quick find that finds what the quick find dialog
finds, and the content menu in compact form, in the model's order, with the
current entry marked. A group of more than six entries SHALL show only its
label and count until opened.

#### Scenario: A large group

- **WHEN** a group holds nine collections
- **THEN** the picker shows its label and count, and clicking the label shows
  the nine

### Requirement: The location picker works by keyboard alone

The picker SHALL open with Enter or Space on the breadcrumb, the arrow keys
SHALL move through its entries, Enter SHALL open the focused one, and Esc
SHALL close the picker and give focus back to the breadcrumb. The breadcrumb
SHALL say whether the picker is open.

#### Scenario: Keyboard only

- **WHEN** somebody opens the picker with Enter, moves with the arrow keys and
  presses Esc
- **THEN** the picker closes and the breadcrumb has focus again

### Requirement: Moving between content screens keeps the screen

Moving between a collection and a page, in either direction, SHALL keep the
project's frame and content menu on the screen. The interface SHALL show what
it last loaded for the project and its menu at once, and replace it with the
fresh answer when that comes. Signing out SHALL forget what was kept.

#### Scenario: From a collection to a page and back

- **WHEN** somebody opens a page from a collection and then goes back to the
  collection
- **THEN** the screen never shows only "Looking…" in between

#### Scenario: Signing out

- **WHEN** somebody signs out and another person signs in on the same window
- **THEN** nothing the first person loaded is shown to the second
