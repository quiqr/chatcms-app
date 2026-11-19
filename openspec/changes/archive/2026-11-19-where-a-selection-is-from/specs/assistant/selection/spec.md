## MODIFIED Requirements

### Requirement: The assistant is told which page a selection came from

For a message with a selection, the assistant SHALL be given the request,
the selected text, its context, and where the text comes from: a field of a
page, shared text, or the site's design. Where the text comes from SHALL be
found in the site's source as the model describes it, for any generator,
starting with the page the preview address maps to where that is known (for
a Hugo site, through Hugo's own list of pages and their addresses). When
nothing can be worked out, the assistant SHALL be told to find the text
through its page tools.

#### Scenario: a page that matches

- **WHEN** the person selects the title on the preview of the home page and
  asks for a change
- **THEN** the assistant SHALL be told it is the field "title" of the page
  "home", and its change SHALL be saved as the person's change to that page

#### Scenario: shared text

- **WHEN** the person selects a footer text held by a single in
  `data/settings.yaml`
- **THEN** the assistant SHALL be told it is shared text that shows on more
  pages, and to say so before changing it

#### Scenario: the design

- **WHEN** the person selects text no page or shared file holds
- **THEN** the assistant SHALL be told it comes from the site's design, which
  it cannot change

#### Scenario: no page matches

- **WHEN** where the text comes from cannot be worked out at all
- **THEN** the assistant SHALL be told no page was found for the address

## ADDED Requirements

### Requirement: Where a selection comes from can be asked

The application SHALL answer, for a selection, whether its text is in a
field of a page (with the page, its title and the field), shared text (with
the single and the field) or part of the site's design, to anybody who may
use the chat in that project.

#### Scenario: asking

- **WHEN** the editor asks where "Korenaar Bakery" on the home page comes from
- **THEN** the answer is the field "title" of "Home page"
