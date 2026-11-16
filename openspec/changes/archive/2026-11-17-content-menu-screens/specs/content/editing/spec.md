# Spec Delta

## ADDED Requirements

### Requirement: An open item knows its place in its collection

With a collection item open, the editor header SHALL lead back to the
collection, say "<collection> · n of N", and step to the previous and next
item in the order, search and folder the list had when the item was opened.
For a single it SHALL name the group the single is in.

#### Scenario: Stepping through

- **WHEN** the 12th of 48 posts is open and somebody presses next, or Alt+Down
- **THEN** the 13th opens and the header says 13 of 48

#### Scenario: The ends

- **WHEN** the first item is open
- **THEN** previous is unavailable, and at the last item next is

#### Scenario: A filtered list

- **WHEN** the list was searched for "spring" and showed 3 items
- **THEN** previous and next step through those 3, and the header says n of 3

#### Scenario: Going back

- **WHEN** somebody presses back
- **THEN** the collection is shown with the search, folder and sort it had

### Requirement: Whether an item shows on the website is a switch

When the model gives an item a `draft` field, the editor SHALL show a switch
"Show on website" at the top of the form, on when the item is not a draft. The
word "draft" and the field's name SHALL NOT be shown to an editor.

#### Scenario: Turning it off

- **WHEN** somebody turns "Show on website" off and the page saves
- **THEN** its front matter holds `draft: true`, exactly as the form's own
  boolean field would write it

#### Scenario: A model without the field

- **WHEN** the model gives the item no `draft` field
- **THEN** no switch is shown
