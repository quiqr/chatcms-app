# preview/pane Specification

## Purpose
Defines how the preview column shows each kind of preview.

## Requirements

### Requirement: The pane picks its viewer by the preview type

The preview column SHALL show a web preview as a framed page with live
reload, a pdf preview in the browser's own PDF viewer, and any other preview
as a link to download it.

#### Scenario: web

- **WHEN** a site's preview type is web
- **THEN** the pane SHALL frame the page

#### Scenario: pdf

- **WHEN** a site's preview type is pdf
- **THEN** the pane SHALL show the PDF it declared

#### Scenario: other

- **WHEN** a site's preview type is other
- **THEN** the pane SHALL offer the preview as a download
