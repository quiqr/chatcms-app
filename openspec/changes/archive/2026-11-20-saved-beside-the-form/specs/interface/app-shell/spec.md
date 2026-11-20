## ADDED Requirements

### Requirement: Saving is said beside the form, not in the top bar

Whether a page is saved SHALL be said only beside the form being edited, as
"Saved", "Saving…" or that it is not saved yet. The top bar SHALL NOT carry a
save status, so nothing there reads as a contradiction of the number of
changes waiting to be published.

#### Scenario: editing a page

- **WHEN** a person's change to a page has been saved
- **THEN** the form SHALL say "Saved"

#### Scenario: any screen in a project

- **WHEN** a person opens any screen in a project, at 1440 px or 390 px
- **THEN** the top bar SHALL NOT say whether anything is saved
