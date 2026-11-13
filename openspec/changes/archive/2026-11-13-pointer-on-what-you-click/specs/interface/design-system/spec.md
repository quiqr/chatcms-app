## ADDED Requirements

### Requirement: What can be clicked shows the hand

Every control that does something when clicked SHALL show the pointing hand
cursor when the mouse is over it, whether it is built as a link or as a
button: buttons, menu items, radio cards, segments, tabs, disclosures and the
labels of checkboxes, radios and file pickers. A control that is disabled
SHALL NOT show the hand.

#### Scenario: A button that is not a link

- **WHEN** the mouse is over the "New project" button
- **THEN** the cursor is the pointing hand, as it is over a link

#### Scenario: A menu item

- **WHEN** the person menu is open and the mouse is over "Your account"
- **THEN** the cursor is the pointing hand

#### Scenario: A disabled control

- **WHEN** a button is disabled
- **THEN** the cursor is not the pointing hand

#### Scenario: The rule is checked

- **WHEN** a component sets the arrow cursor on itself
- **THEN** the frontend check fails and names it
