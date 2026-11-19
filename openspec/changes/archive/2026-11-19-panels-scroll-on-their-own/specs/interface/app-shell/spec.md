# Spec Delta

## MODIFIED Requirements

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
