# Spec Delta

## ADDED Requirements

### Requirement: A site that is not connected is not published

Publishing a site kept by ChatCMS SHALL be refused with a message saying it
is not connected yet and where to connect it. The refusal SHALL NOT be
recorded as a failed publish.

#### Scenario: Publishing before connecting

- **WHEN** somebody allowed to publish publishes a site kept by ChatCMS
- **THEN** it SHALL be refused, naming project settings as the place to
  connect it
- **AND** no publish SHALL be recorded

#### Scenario: Publishing after connecting

- **WHEN** the same site has been connected to a repository
- **THEN** publishing SHALL send its work there as for any other site
