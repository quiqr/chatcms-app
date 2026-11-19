# Spec Delta

## ADDED Requirements

### Requirement: Without an origin of its own, a preview follows the app's address

Without a preview origin, a preview SHALL be reached through the address the
browser reached the app by, whatever name that is, so the person's session
reaches it.

#### Scenario: the app reached by another name

- **WHEN** the app's base URL is on `127.0.0.1` and a person uses it through
  `localhost`
- **THEN** their preview SHALL show, and a selection in it SHALL reach the
  app
