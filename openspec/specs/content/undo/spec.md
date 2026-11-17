# content/undo Specification

## Purpose
Lets somebody take back edits to the page they have open, step by step or back
to a point they pick from a list, without fear of losing anything, since every
edit is saved at once.

## Requirements

### Requirement: Each save is one undo step

Each save the editor makes for the open page SHALL become one undo step,
holding every field and the page text it changed, with the value before and
after, and the time it was made. Only a save that succeeded SHALL make a step.

#### Scenario: Two edits in one pause

- **WHEN** somebody changes the title and then the summary without stopping,
  and the editor saves once
- **THEN** there is one step, naming the title and the summary

#### Scenario: A refused save

- **WHEN** a save is refused
- **THEN** no step is made and there is nothing to undo

### Requirement: Undo and redo are a save

Undoing a step SHALL write the values from before it, and redoing it the
values after it, through the same save as any edit, so that each makes a
version. Undo SHALL first save anything still waiting to be written, as a step
of its own, and then undo that step.

#### Scenario: Undoing a cleared field

- **WHEN** somebody clears the summary, it is saved, and they press Undo
- **THEN** the summary holds its old text again
- **AND** that is saved as a new version

#### Scenario: Redo

- **WHEN** somebody presses Undo and then Redo
- **THEN** the page is as it was before the undo, and saved so

#### Scenario: Undo before the save

- **WHEN** somebody types in the title and presses Undo before the editor has
  saved
- **THEN** the typing is saved and then undone, and the title is as it was

### Requirement: Undo touches only its own fields

An undo or redo SHALL write only the fields of its steps. A field that no
longer holds the value the step left in it SHALL be left as it is, and the
editor SHALL say which field it left alone.

#### Scenario: A field changed since

- **WHEN** a step changed the title and the summary, and the summary no longer
  holds what that step left in it
- **THEN** Undo puts back the title only
- **AND** the editor says the summary was left as it is

### Requirement: Undo and redo are always in reach

Undo and redo buttons SHALL be next to whether the page is saved, in the
editor at every width, each disabled when there is nothing to undo or redo,
and each saying in its name what it would undo or redo.
Ctrl+Z or Cmd+Z SHALL undo, and Ctrl+Shift+Z, Cmd+Shift+Z or Ctrl+Y redo, when
the focus is not in a text field. Inside a text field the keys SHALL keep the
browser's own meaning.

#### Scenario: Nothing to undo yet

- **WHEN** a page has just been opened
- **THEN** Undo and Redo are both disabled

#### Scenario: The keys inside a text field

- **WHEN** the focus is in the body text and somebody presses Ctrl+Z
- **THEN** the browser's undo inside that field happens, not the page's

#### Scenario: On a phone

- **WHEN** the editor is shown 390 px wide
- **THEN** Undo, Redo and Undo history are there, each a target of at least
  44 px

### Requirement: The undo history can be seen and used

An "Undo history" button beside Undo and Redo SHALL open a list of the steps
since the page was opened,
newest first, each naming the fields it changed as the model names them, and
its time, with "Opened" as the last row and the current point marked. Picking
a row SHALL bring the page back to that point in one save; the later steps
SHALL stay available as redo. The list SHALL never show a version identifier
or a git word.

#### Scenario: Going back three steps at once

- **WHEN** there are three steps and somebody picks "Opened"
- **THEN** the page is as it was when it was opened, after one save
- **AND** the three steps are listed as ones that can be redone

#### Scenario: A row names the fields

- **WHEN** a step changed the title and the page text
- **THEN** its row says "Title, Page text" and its time

### Requirement: A new edit ends the redo

An edit after an undo SHALL drop the steps that could have been redone.

#### Scenario: Typing after an undo

- **WHEN** somebody undoes a step and then changes a field
- **THEN** Redo is disabled and the undone step is no longer listed

### Requirement: The steps last as long as the tab

The steps SHALL be kept for the browser tab, per person, project and page, and
SHALL survive a reload of that tab. At most 100 steps SHALL be kept, dropping
the oldest. When the page as loaded no longer matches where the steps left it,
they SHALL be thrown away.

#### Scenario: A reload

- **WHEN** somebody makes two steps and reloads the tab
- **THEN** both steps are listed and Undo works

#### Scenario: Changed elsewhere

- **WHEN** the page was changed by somebody else between leaving it and
  opening it again in the same tab
- **THEN** the editor starts with no steps

#### Scenario: Another page

- **WHEN** somebody opens a different page
- **THEN** that page has its own steps, and the first page's steps are kept
  for when they come back to it
