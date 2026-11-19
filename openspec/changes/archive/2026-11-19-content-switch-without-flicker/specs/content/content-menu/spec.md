## ADDED Requirements

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
