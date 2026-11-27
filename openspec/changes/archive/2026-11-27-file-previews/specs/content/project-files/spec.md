## MODIFIED Requirements

### Requirement: A file opens in a code editor

Opening a text file SHALL show it in a code editor with line numbers,
search and highlighting by its kind; Go template actions in HTML SHALL be
marked. An image, video, audio file or PDF SHALL be shown the way the
browser shows it. Any other binary file, or a text file over 1 MB, SHALL say
it cannot be opened here.

#### Scenario: a template

- **WHEN** an implementer opens `layouts/index.html`
- **THEN** its text SHALL be shown with line numbers

#### Scenario: a picture

- **WHEN** an implementer opens a PNG
- **THEN** the picture SHALL be shown, with a link that opens it in a new tab

#### Scenario: a film

- **WHEN** an implementer opens an MP4
- **THEN** it SHALL be shown with the browser's controls for playing it

#### Scenario: a document

- **WHEN** an implementer opens a PDF
- **THEN** it SHALL be shown in a frame

#### Scenario: an archive

- **WHEN** an implementer opens a ZIP
- **THEN** the screen SHALL say this file cannot be opened here

## ADDED Requirements

### Requirement: Only a file of a known kind is served as itself

The server SHALL serve a file's bytes only for the kinds it can preview,
chosen by extension, with that kind's content type and
`X-Content-Type-Options: nosniff`, and SHALL serve ranges of it. Every
other file SHALL be refused. The path rules and the permission SHALL be
those of reading a file.

#### Scenario: a picture

- **WHEN** an implementer asks for the bytes of `static/images/loaf.png`
- **THEN** the answer SHALL be `image/png` with `nosniff`

#### Scenario: a template

- **WHEN** an implementer asks for the bytes of `layouts/index.html`
- **THEN** the request SHALL be refused

#### Scenario: an editor

- **WHEN** an editor asks for the bytes of a picture
- **THEN** the request SHALL not be allowed
