# preview/pane Specification

## Purpose
Defines how the preview column shows each kind of preview.

## Requirements

### Requirement: The pane picks its viewer by the preview type

The preview column SHALL show a web preview as a framed page with live
reload, a pdf preview in a PDF.js viewer served under the preview's own
address, and any other preview as a link to download it. The PDF viewer
SHALL keep the page it showed when the preview is refreshed, and SHALL let
the person select text and send it to the assistant like a web preview.

#### Scenario: web

- **WHEN** a site's preview type is web
- **THEN** the pane SHALL frame the page

#### Scenario: pdf

- **WHEN** a site's preview type is pdf
- **THEN** the pane SHALL show the PDF it declared, with selectable text

#### Scenario: pdf refreshed

- **WHEN** the person refreshes a pdf preview while on its third page
- **THEN** the viewer SHALL show the third page again

#### Scenario: other

- **WHEN** a site's preview type is other
- **THEN** the pane SHALL offer the preview as a download

### Requirement: A PDF preview keeps its own size

A pdf preview SHALL start with its pages at their own size, one PDF point to
one CSS point, whatever the width of the pane, and SHALL scroll where the
pane is smaller. On a phone it SHALL instead start with each page scaled down
to the width of the screen less a margin, keeping its proportions, and never
past its own size. The pages SHALL stand on a grey ground, so that the edges
of the page format show. A different size SHALL only come from the viewer's
zoom, which keeps the page's proportions; the pane SHALL NOT reflow a PDF to
its width.

#### Scenario: a narrow pane

- **WHEN** an A5 PDF is previewed in a pane narrower than the page
- **THEN** the page SHALL keep its size and the pane SHALL scroll

#### Scenario: on a phone

- **WHEN** a PDF is previewed on a phone 390 px wide
- **THEN** the whole width of the page SHALL show, with the grey ground
  around it, and its proportions SHALL be those of the PDF

### Requirement: The preview follows what an item is

Beside a form, the preview SHALL show the address the model declares for a
single (`previewUrl`), nothing where the model hides the preview
(`hidePreviewIcon`), the item's own page when its file is a page, and, when
its file is data (JSON, YAML or TOML), an explanation that it has no page of
its own with a way to show the home page instead.

#### Scenario: data without a page

- **WHEN** a person edits a single whose file is `data/hours.json`
- **THEN** the preview SHALL say "No page of its own" and why
- **AND** "Show the home page" SHALL show the site's home page

#### Scenario: a declared address

- **WHEN** a single declares `previewUrl: /about/`
- **THEN** the preview SHALL show `/about/`

#### Scenario: a hidden preview

- **WHEN** a single declares `hidePreviewIcon: true`
- **THEN** the preview SHALL say this site shows no preview for it

### Requirement: The preview catches up after an assistant's change

After the assistant changes a page, the preview SHALL load the page again at
once and twice more within the following seconds, so a site that was still
rebuilding when the first load came is shown with the change.

#### Scenario: a slow rebuild

- **WHEN** the assistant changes a page while the machine is busy
- **THEN** the preview SHALL show the new text within ten seconds
