## ADDED Requirements

### Requirement: A symbol says how the workspace stands

Beside the save status, the editor, project and phone shells SHALL show a
symbol for the person's own workspace, telling apart starting, ready,
sleeping and did not start.

#### Scenario: ready

- **WHEN** a person's workspace is warm
- **THEN** the symbol SHALL say the preview is ready

#### Scenario: asleep

- **WHEN** a person's workspace sleeps
- **THEN** the symbol SHALL say the preview is asleep

### Requirement: Clicking the symbol says more, in plain words

Clicking the symbol SHALL open a popover that says the state of the preview,
since when, and whether the assistant can be asked, without git words or
backend names for somebody who may not see technical detail.

#### Scenario: an editor

- **WHEN** an editor opens the popover of a warm workspace
- **THEN** it SHALL say the preview is ready and since when
- **AND** it SHALL say whether the assistant can be asked
- **AND** it SHALL NOT name the backend, show the log or offer to start again

### Requirement: Implementers see the backend and can start again

For somebody who may see a project's technical detail, the popover SHALL
also name what the workspace runs on, list the start steps and the log, and
offer to start the workspace again, which puts it to sleep and opens it.

#### Scenario: an implementer starts again

- **WHEN** an implementer presses Start again
- **THEN** the workspace SHALL be put to sleep and opened again
- **AND** the symbol SHALL say it is starting

#### Scenario: somebody else asks to start again

- **WHEN** somebody who may not see technical detail asks the server to
  start a workspace again
- **THEN** the server SHALL refuse with 403
