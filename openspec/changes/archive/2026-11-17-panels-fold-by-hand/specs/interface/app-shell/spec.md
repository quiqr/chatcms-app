## REMOVED Requirements

### Requirement: The editor frame leaves room for what comes later

**Reason**: it made the editor swap the project navigation for a rail by
itself, which people read as the screen jumping.
**Migration**: replaced by "The editor frame keeps the panels as they were
left" and "The project navigation folds by hand".

## ADDED Requirements

### Requirement: The editor frame keeps the panels as they were left

The editor SHALL lay out as a row of slots: the project navigation, the
content column, a slot for the form and a slot for the preview. The editor
SHALL show the project navigation and the content column in the state the
person left them, the same as every other screen inside the project.

#### Scenario: The editor has its slots

- **WHEN** the editor frame is rendered
- **THEN** it has the project navigation, the content column, a slot for the
  form and a slot for the preview

#### Scenario: Opening an item leaves the navigation alone

- **WHEN** the project navigation is unfolded and somebody opens an item from
  a collection
- **THEN** the editor shows the project navigation unfolded
- **AND** going back to the collection shows it unfolded too

#### Scenario: The content column folded

- **WHEN** the content column is folded in the editor
- **THEN** the form and preview take its room
- **AND** the column's own unfold icon brings it back

### Requirement: The project navigation folds by hand

On screens 640 px and wider, the project navigation SHALL have an icon in its
heading that folds it to a narrow rail. The rail SHALL hold the same entries
as icons, each with an accessible name and a tooltip, and an icon at its top
that unfolds it. The state SHALL be remembered per person, across projects
and reloads, and SHALL change only when the person folds or unfolds it.

#### Scenario: Folding the navigation

- **WHEN** somebody folds the project navigation on the collection view
- **THEN** it becomes a rail of icons with the same entries
- **AND** the setup entries are there for an owner or implementer, as when
  unfolded

#### Scenario: The fold is kept

- **WHEN** somebody has folded the project navigation and opens an item,
  reloads, or switches to another project
- **THEN** the navigation is still folded

#### Scenario: Unfolding the navigation

- **WHEN** somebody presses the unfold icon at the top of the rail
- **THEN** the full navigation is shown again

### Requirement: A panel moves when the person folds it

Folding or unfolding a side panel SHALL show a short movement of its width,
taking no more than 200 ms. A screen that loads or reloads SHALL show the
panels in their remembered state without movement. With reduced motion asked
for, a panel SHALL change at once. The movement SHALL NOT delay the result:
the folded or unfolded panel is usable as soon as it is pressed.

#### Scenario: Reduced motion

- **WHEN** the system asks for reduced motion and somebody folds a panel
- **THEN** the panel changes without movement

#### Scenario: A reload

- **WHEN** a screen loads with a panel remembered as folded
- **THEN** the panel is shown folded from the start, without moving
