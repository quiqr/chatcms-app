## ADDED Requirements

### Requirement: Signing out always says so

After a person signs out, the sign-in screen SHALL say "You have signed out",
however many requests were still on their way when the session ended; a
request refused because the session is gone SHALL NOT send them to the
sign-in screen by the route that does not say it.

#### Scenario: a request still on its way

- **WHEN** a person signs out while the workspace symbol is asking how the
  preview stands
- **THEN** the sign-in screen SHALL say "You have signed out"
