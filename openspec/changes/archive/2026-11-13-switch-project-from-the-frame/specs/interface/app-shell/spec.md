## ADDED Requirements

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
