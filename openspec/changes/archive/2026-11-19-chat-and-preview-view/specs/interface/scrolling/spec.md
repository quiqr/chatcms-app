## MODIFIED Requirements

### Requirement: The editor shows what fits and lets the person choose the rest

The editor SHALL offer four views at every desktop width: Form, Form +
preview, Chat + preview and Preview, and SHALL remember the choice per
person per project. Form + preview SHALL be offered only where form and
preview both keep their minimum width; where they do not, the form SHALL be
shown. Chat + preview SHALL show a chat column of 400 px beside the preview,
and the chat alone where there is no room for both. Views SHALL NOT be
stacked one under the other.

#### Scenario: a medium window

- **WHEN** the editor is opened in a window too narrow for form and preview
  side by side
- **THEN** the form is shown, and the preview is one choice away

#### Scenario: a wide window

- **WHEN** the window fits both
- **THEN** form and preview are shown side by side

#### Scenario: chat and preview

- **WHEN** somebody chooses Chat + preview
- **THEN** the chat shows on the left at 400 px and the preview beside it,
  without the form
- **AND** the next page they open in that project opens in Chat + preview

## REMOVED Requirements

### Requirement: The assistant takes the form's place when there is no room

**Reason**: The assistant no longer opens beside other views; it has its own
view, Chat + preview.
**Migration**: Choose Chat + preview in the view switch.
