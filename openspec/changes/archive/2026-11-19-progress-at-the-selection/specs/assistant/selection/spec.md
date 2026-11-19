## ADDED Requirements

### Requirement: A sent request shows its progress at the selection

After a request is sent from the selection, a pill SHALL stand where the
input was, following the text it is about. While the request is being
answered it SHALL say it is working and offer Stop. When the answer changed
the page it SHALL say "Changed" with Undo and Show in chat, and the changed
text SHALL be tinted for a moment; the pill SHALL stay at least 8 seconds and
until the person touches the preview again. When the answer is a question it
SHALL say so and offer Answer in chat. Every request SHALL have a pill of its
own, and pills SHALL survive a reload of the preview.

#### Scenario: working, then done

- **WHEN** somebody sends "Make it shorter" from a selection
- **THEN** a pill says "Working on it…" with Stop
- **AND** once the page is changed it says "Changed" with Undo and Show in
  chat, and the new text is tinted

#### Scenario: undo at the selection

- **WHEN** somebody presses Undo on that pill
- **THEN** the page holds what it held before, the chat says the change was
  undone, and the pill goes

#### Scenario: a question

- **WHEN** the assistant answers with a question and changes nothing
- **THEN** the pill says the assistant has a question, and Answer in chat
  opens Chat + preview

#### Scenario: two at once

- **WHEN** somebody sends two requests from two selections
- **THEN** each has its own pill, and Chat + preview counts 2

#### Scenario: stop

- **WHEN** somebody presses Stop on a working pill
- **THEN** the assistant stops answering and the pill goes
