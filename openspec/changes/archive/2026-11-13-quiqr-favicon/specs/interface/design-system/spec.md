## ADDED Requirements

### Requirement: The browser tab carries the Quiqr icon

The application SHALL name the Quiqr icon, the same one quiqr.org uses, as
its icon, and SHALL serve it itself, so that a browser tab, a bookmark and a
phone's home screen show it.

#### Scenario: A tab

- **WHEN** a screen of the application is open in a browser
- **THEN** its tab shows the Quiqr icon

#### Scenario: The build carries it

- **WHEN** the frontend is built for production
- **THEN** the page names the icon and the icon is in the build, or the
  frontend check fails
