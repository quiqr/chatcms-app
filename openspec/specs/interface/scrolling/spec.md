# interface/scrolling Specification

## Purpose
Defines what scrolls in a desktop frame and what stays in view.

## Requirements

### Requirement: The page does not scroll; each panel scrolls on its own

A desktop frame SHALL be exactly as tall as the window. The page itself SHALL
NOT scroll. The header SHALL stay at the top. Each panel (the navigation, the
content menu, the main area, the form, the preview and the assistant) SHALL
scroll inside itself when its content is taller than the panel.

#### Scenario: a long form

- **WHEN** a page with more fields than fit the window is edited
- **THEN** only the form scrolls, and the header, the preview and the
  assistant stay where they are

#### Scenario: a long list

- **WHEN** a screen inside a project shows a list taller than the window
- **THEN** the list scrolls inside the main area, and the header and the
  navigation stay in view

### Requirement: The preview and the assistant fill the height they are given

The preview SHALL fill the height of its panel, whatever the height of the
form. The assistant's message box SHALL stay at the bottom of its panel and
in view, however long the conversation is.

#### Scenario: the end of the preview

- **WHEN** the preview is shown beside a form taller than the window
- **THEN** the bottom of the preview is in view without scrolling the page

#### Scenario: the message box

- **WHEN** the assistant is open
- **THEN** its message box is in view, at any length of form or conversation

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
