## ADDED Requirements

### Requirement: Every frame has a top bar with the person menu at its end

Every frame SHALL have a top bar across the top of its content, and the person
menu SHALL be the last thing in it, so that the way to yourself and the way out
are in the same corner on every screen. No frame SHALL place the person menu
in its side navigation.

#### Scenario: Inside a project

- **WHEN** a screen inside a project is shown on a desktop
- **THEN** the person menu is in the top bar, after the publish action
- **AND** the project navigation holds no person menu

#### Scenario: In an admin area

- **WHEN** the organisation admin area or the platform admin area is shown
- **THEN** it has a top bar holding the way back to one's own projects and,
  at its end, the person menu
- **AND** the top bar is light, as it is everywhere else, even where the side
  navigation is dark

#### Scenario: Every other frame

- **WHEN** the organisation frame, the editor frame or the phone frame is
  shown
- **THEN** the person menu is at the end of its top bar

### Requirement: The person menu says who you are here

The person menu SHALL name the viewer and, where the frame knows it, the role
they have there.

#### Scenario: A role on a project

- **WHEN** an editor opens the person menu inside a project
- **THEN** it shows their name and that they are an editor on it

#### Scenario: An admin area

- **WHEN** the person menu is opened in the platform admin area
- **THEN** it shows their name and "Platform admin"
