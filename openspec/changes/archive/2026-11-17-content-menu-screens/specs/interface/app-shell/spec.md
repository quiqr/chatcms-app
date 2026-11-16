# Spec Delta

## MODIFIED Requirements

### Requirement: The editor frame leaves room for what comes later

The editor SHALL lay out as a row of slots with a narrow icon rail. The slot
before the form SHALL hold the content column, which can be folded away.

#### Scenario: The editor has its slots

- **WHEN** the editor frame is rendered
- **THEN** it has the content column, a slot for the form and a slot for the
  preview

#### Scenario: The rail replaces the full navigation

- **WHEN** the editor is shown
- **THEN** the project navigation is a narrow rail rather than the full
  sidebar

#### Scenario: The content column folded

- **WHEN** the content column is folded
- **THEN** the form and preview take its room
- **AND** the rail's Content icon brings it back
