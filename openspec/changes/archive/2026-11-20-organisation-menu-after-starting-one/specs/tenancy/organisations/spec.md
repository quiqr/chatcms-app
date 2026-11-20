## ADDED Requirements

### Requirement: A new organisation shows in the organisation menu at once

When a platform admin starts an organisation, who is signed in SHALL be read
again, so an organisation the admin belongs to appears in the organisation
menu without a reload.

#### Scenario: starting one for oneself

- **WHEN** a platform admin starts an organisation with themselves as owner
- **THEN** the organisation menu SHALL list it without a reload
