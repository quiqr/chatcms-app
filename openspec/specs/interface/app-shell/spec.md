# interface/app-shell Specification

## Purpose
Defines the frames every screen sits in, and how each behaves as the window
narrows, so that working on a phone is the same product rather than a reduced
one.

## Requirements

### Requirement: A screen inside a project has the project frame

A screen inside a project SHALL sit in a frame carrying the project's
navigation, the current sync state, and the publish action with a count of
what is unpublished.

#### Scenario: The frame carries navigation

- **WHEN** a screen inside a project is shown
- **THEN** the frame offers the site's sections and, for an owner or
  implementer, the setup sections

#### Scenario: Setup is hidden, not disabled

- **WHEN** the person is an editor or a viewer
- **THEN** the setup sections are absent rather than shown greyed out

#### Scenario: The publish action carries its count

- **WHEN** there are unpublished changes
- **THEN** the publish action shows how many

### Requirement: The frame reflows rather than breaking

The frame SHALL reflow as the window narrows: the navigation gives way to the
content, and below the phone threshold the phone frame is used instead.

#### Scenario: A narrow window

- **WHEN** the window is too narrow for navigation beside content
- **THEN** the navigation moves above the content rather than squeezing it

#### Scenario: A phone

- **WHEN** the window is at or below the phone threshold
- **THEN** the phone frame is used, with its bottom tab bar
- **AND** the project navigation frame is not shown as well

#### Scenario: No sideways scrolling

- **WHEN** any screen is shown at the phone width
- **THEN** the page does not scroll sideways

### Requirement: The phone frame respects the device

The phone frame SHALL keep its actions reachable and clear of the device's own
furniture.

#### Scenario: The bottom bar clears the device

- **WHEN** the phone frame is shown on a device with a home indicator
- **THEN** the bottom bar is padded clear of it

#### Scenario: The editor swaps the tab bar for its actions

- **WHEN** a page is being edited on a phone
- **THEN** the tab bar gives way to the editing actions, including publish

### Requirement: Organisation screens have their own frame

A screen above the level of a single project SHALL use a header rather than
the project navigation.

#### Scenario: The project list

- **WHEN** the list of projects is shown
- **THEN** it sits under a header with the organisation switcher, not inside a
  project's navigation

### Requirement: The editor frame leaves room for what comes later

The editor SHALL lay out as a row of slots with a narrow icon rail, so that a
third panel can be added without the layout being rebuilt.

#### Scenario: The editor has its slots

- **WHEN** the editor frame is rendered
- **THEN** it has a slot for the form, a slot for the preview, and a slot that
  is empty for now

#### Scenario: The rail replaces the full navigation

- **WHEN** the editor is shown
- **THEN** the project navigation is a narrow rail rather than the full
  sidebar

### Requirement: Every screen of this milestone is reachable

Each screen the milestone names SHALL have an address, so that a screen can be
built without navigation being invented alongside it.

#### Scenario: A screen has an address

- **WHEN** any milestone screen is named
- **THEN** there is an address that reaches it

#### Scenario: An address that reaches nothing

- **WHEN** an address is used that names no screen
- **THEN** the person is told so in plain language and offered a way back

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
