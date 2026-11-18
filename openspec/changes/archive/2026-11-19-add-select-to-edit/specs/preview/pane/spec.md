# Spec Delta

## MODIFIED Requirements

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
