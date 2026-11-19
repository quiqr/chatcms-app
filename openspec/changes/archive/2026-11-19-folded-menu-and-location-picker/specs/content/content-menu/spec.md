# Spec Delta

## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: A location picker reaches any page without the column

On screens wider than a phone, the collection view and the editor SHALL show a
breadcrumb `Content / <group> / <current>`, whose last part opens a picker,
whether the column is folded or not. Choosing an entry in the picker SHALL
open it and close the picker. "Show the full content menu" SHALL unfold the
column.

#### Scenario: Switching while folded

- **WHEN** the column is folded and somebody opens the picker and chooses
  News
- **THEN** News opens and the picker closes, and the column stays folded

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
