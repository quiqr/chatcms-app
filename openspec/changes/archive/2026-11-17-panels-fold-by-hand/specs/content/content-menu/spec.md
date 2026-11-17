## RENAMED Requirements

- FROM: `### Requirement: The content column folds away beside the editor`
- TO: `### Requirement: The content column folds by hand`

## MODIFIED Requirements

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
