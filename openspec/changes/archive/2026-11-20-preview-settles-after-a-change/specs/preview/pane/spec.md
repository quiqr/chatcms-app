## ADDED Requirements

### Requirement: The preview catches up after an assistant's change

After the assistant changes a page, the preview SHALL load the page again at
once and twice more within the following seconds, so a site that was still
rebuilding when the first load came is shown with the change.

#### Scenario: a slow rebuild

- **WHEN** the assistant changes a page while the machine is busy
- **THEN** the preview SHALL show the new text within ten seconds
