## ADDED Requirements

### Requirement: The publish count follows changes

The number of changes waiting to be published, on the Publish control and in
the project navigation, SHALL follow a save, a new or removed page, or an
assistant's change without the page being opened again.

#### Scenario: a save

- **WHEN** a person saves a change to a page while everything was published
- **THEN** the Publish control SHALL show one change waiting, without a reload
