## ADDED Requirements

### Requirement: A page that came out empty is loaded again

When a page of the site arrives in the preview with no text and no picture,
the preview SHALL load it again after a short wait, at most three times; a
page the app draws itself, such as the PDF viewer, SHALL never count as
empty.

#### Scenario: a page asked for while the site was still being built

- **WHEN** the preview receives an empty page
- **THEN** it SHALL load the page again, and show it once it has content

#### Scenario: the PDF viewer before its file arrives

- **WHEN** the PDF viewer is shown before its file has loaded
- **THEN** the preview SHALL NOT load it again
