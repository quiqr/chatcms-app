# Spec Delta

## Purpose

Defines what the agent may do to a site, and that it does so through the same
rules as a person using the forms.

## ADDED Requirements

### Requirement: The agent changes pages only through the host's tools

The agent SHALL list pages, read a page and change a page's fields and text
only through tools the host serves, which validate against the Quiqr model
and save exactly as the forms do. Its own tools for writing files, running
commands and fetching from the web SHALL be denied.

#### Scenario: changing a field

- **WHEN** the agent asks to set a page's title
- **THEN** the page SHALL be saved as a version on the person's copy, written
  by that person

#### Scenario: a field the model does not have

- **WHEN** the agent asks to set a field the page's model does not name
- **THEN** it SHALL be refused, and the page left as it was

#### Scenario: writing a file directly

- **WHEN** the agent tries to write a file or run a command itself
- **THEN** it SHALL be denied

### Requirement: The tools act as the workspace's person

The tools SHALL act with the rights of the person whose workspace asked, and
SHALL accept only a live workspace token.

#### Scenario: a viewer's workspace

- **WHEN** the agent of somebody who may only view asks to change a page
- **THEN** it SHALL be refused

#### Scenario: no token

- **WHEN** the tools are called without a live workspace token
- **THEN** they SHALL be refused
