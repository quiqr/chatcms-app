## ADDED Requirements

### Requirement: The viewer shows which page is in view and goes to any page

The PDF viewer SHALL show a toolbar above the pages that stays in place while
the pages scroll. The toolbar SHALL show the number of the page in view and
the number of pages, SHALL offer previous and next, and SHALL go to the page
whose number is typed. A number outside the PDF SHALL go to the nearest page.

#### Scenario: Going to a page

- **WHEN** somebody types 7 in the page field of a 12-page PDF and presses
  Enter
- **THEN** page 7 is in view and the toolbar says "7 / 12"

#### Scenario: Previous at the first page

- **WHEN** the first page is in view
- **THEN** previous is unavailable

### Requirement: The viewer zooms without changing the page's format

The viewer SHALL zoom out and in by steps and SHALL go back to 100%, the
page's own size, in one press. Zooming SHALL keep the page's proportions and
the page in view, and text SHALL stay selectable at every zoom.

#### Scenario: Zooming in

- **WHEN** somebody zooms in from 100%
- **THEN** the pages are drawn at 125%, the same page stays in view and the
  toolbar says 125%

#### Scenario: Back to the own size

- **WHEN** somebody presses 100% after zooming
- **THEN** the pages are at their own size again

### Requirement: The viewer offers the PDF's contents

When the PDF has an outline, the viewer SHALL offer a contents panel listing
it, nested as in the PDF, and choosing an entry SHALL bring its page into
view. On a phone the panel SHALL cover the pages and close after a choice.
Without an outline the viewer SHALL NOT offer contents.

#### Scenario: Going to a chapter

- **WHEN** somebody opens the contents of a book and chooses "Bread"
- **THEN** the page where "Bread" starts is in view

#### Scenario: No outline

- **WHEN** the PDF has no bookmarks
- **THEN** the toolbar has no contents button

### Requirement: The viewer offers the PDF as a download

The toolbar SHALL offer the PDF as a download under its own file name.

#### Scenario: Downloading

- **WHEN** somebody presses Download
- **THEN** the browser saves the previewed PDF

### Requirement: The viewer speaks the interface's language

Every label in the viewer SHALL come from the interface's catalogue, in the
language the interface is in.

#### Scenario: Dutch

- **WHEN** the interface is in Dutch and a PDF is previewed
- **THEN** the viewer's buttons are labelled in Dutch
