# Spec Delta

## Purpose

Defines what scrolls in a desktop frame and what stays in view.

## ADDED Requirements

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

The editor SHALL show the form and the preview side by side only where both
keep their minimum width. Where they do not, it SHALL show one of them, the
form first, and the choice of form, both or preview SHALL be offered at
every desktop width. They SHALL NOT be stacked one under the other.

#### Scenario: a medium window

- **WHEN** the editor is opened in a window too narrow for form and preview
  side by side
- **THEN** the form is shown, and the preview is one choice away

#### Scenario: a wide window

- **WHEN** the window fits both
- **THEN** form and preview are shown side by side

### Requirement: The assistant takes the form's place when there is no room

Where there is no room for the assistant beside the form and the preview, the
assistant SHALL take the form's place while it is open, so the preview stays
beside it, and the form SHALL come back when it closes. Where there is not
even room for the assistant beside the preview, the assistant SHALL show on
its own.

#### Scenario: opening the assistant on a medium window

- **WHEN** the assistant is opened where it does not fit as a column of its
  own
- **THEN** it shows beside the preview, in the form's place, and closing it
  shows the form again
