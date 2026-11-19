# Spec Delta

## ADDED Requirements

### Requirement: The provider is told who is asking, as it asks to be

Every request the gateway makes SHALL carry a user agent of ChatCMS's own.
When the installation names a session header, every request SHALL carry in
it a value that is the same for one person and project and differs between
them, and that reveals neither.

#### Scenario: OpenCode Go

- **WHEN** the session header is `x-opencode-session` and one person asks
  twice in one project
- **THEN** both requests SHALL carry the same session, and another person's
  requests another

#### Scenario: no session header named

- **WHEN** no session header is named
- **THEN** no session header SHALL be sent

### Requirement: Every call is logged

The gateway SHALL log each call to the provider with its model, its tokens
in and out, and its status.

#### Scenario: a call

- **WHEN** a workspace's request is answered
- **THEN** the log SHALL hold a line saying the model and the tokens used
