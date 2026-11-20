## ADDED Requirements

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
