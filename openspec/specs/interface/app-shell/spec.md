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

The frame SHALL reflow as the window narrows: when it is too narrow for the
panels beside the content, the navigation and the content menu show folded,
as their rail and strip, rather than moving above the content; below the
phone threshold the phone frame is used instead. Showing a panel folded
because the window is narrow SHALL NOT change the fold the person chose.

#### Scenario: A narrow window

- **WHEN** the window is too narrow for navigation beside content
- **THEN** the navigation shows as its rail and the content keeps the room
- **AND** unfolding it shows it over the content until something is picked

#### Scenario: The person's own fold is kept

- **WHEN** a person who keeps the navigation unfolded narrows the window and
  widens it again
- **THEN** the navigation is unfolded again

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

### Requirement: The project switcher switches projects

The project shown at the top of a project's frame SHALL open a menu of the
projects in the same organisation that the viewer can reach, with the current
one marked, and a way to the overview of all projects at its end. Choosing a
project SHALL open that project's content. The phone frame SHALL offer the
same menu from the project in its header.

#### Scenario: Switching to another project

- **WHEN** somebody inside one project opens the project switcher and chooses
  another project
- **THEN** that project's content is shown

#### Scenario: The current project is marked

- **WHEN** the project switcher is open
- **THEN** the project being looked at is marked as the current one

#### Scenario: The way to every project

- **WHEN** the project switcher is open
- **THEN** its last item leads to the overview of all projects

#### Scenario: Archived projects are left out

- **WHEN** a project in the organisation is archived
- **THEN** it is not offered in the project switcher

#### Scenario: On a phone

- **WHEN** a project is shown in the phone frame and the project in its header
  is pressed
- **THEN** the same menu opens

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

### Requirement: The publish count follows changes

The number of changes waiting to be published, on the Publish control and in
the project navigation, SHALL follow a save, a new or removed page, or an
assistant's change without the page being opened again.

#### Scenario: a save

- **WHEN** a person saves a change to a page while everything was published
- **THEN** the Publish control SHALL show one change waiting, without a reload
