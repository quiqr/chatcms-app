# Spec Delta

## Purpose

Defines the script ChatCMS adds to a preview, what passes between the
preview and the app, and the highlight after a change.

## ADDED Requirements

### Requirement: Every web preview page carries the selection script

The preview proxy SHALL add one script element, served under the preview's
own address, to every HTML page it passes on, and SHALL leave every other
answer as the workspace gave it.

#### Scenario: an HTML page

- **WHEN** the workspace answers a preview request with an HTML page
- **THEN** the page SHALL reach the browser with the selection script added
  before its closing body tag, or at its end when it has none

#### Scenario: anything else

- **WHEN** the workspace answers with a stylesheet, a picture or a PDF
- **THEN** the answer SHALL pass unchanged

### Requirement: A selection reaches the app, and only the app

When the person selects text in a preview and presses the button the script
shows, the script SHALL post the selected text, up to 200 characters before
and after it, and the page's address to the parent window, addressed to the
app's origin only. The app SHALL take such a message only from its preview
frame and from the preview's origin.

#### Scenario: selecting a sentence

- **WHEN** the person selects a sentence in the preview and presses the
  button
- **THEN** the app SHALL receive the sentence, its context and the page's
  address

#### Scenario: a message from elsewhere

- **WHEN** a message shaped like a selection comes from another window or
  another origin
- **THEN** the app SHALL ignore it

### Requirement: The changed text is highlighted after the assistant's change

After a change the assistant made for a selection, the reloaded preview
SHALL highlight, for a few seconds, the text that now stands between the
selection's old context.

#### Scenario: a changed sentence

- **WHEN** the assistant changed the selected sentence and the preview
  reloads
- **THEN** the new sentence SHALL be highlighted briefly
