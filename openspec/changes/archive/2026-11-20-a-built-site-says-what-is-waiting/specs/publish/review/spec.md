## MODIFIED Requirements

### Requirement: What changed is said as pages, not as files

The review SHALL list what differs from the published site as pages and
sections, using the names the site's model gives them. For a site that
publishes built pages, the site as last published is where its own files stood
when those pages were published, because its own files never reach the remote.

#### Scenario: a page the model knows

- **WHEN** a file the model names as a page has changed
- **THEN** it SHALL be listed under the title the model gives it

#### Scenario: a page in a collection

- **WHEN** a file inside a collection has changed
- **THEN** it SHALL be listed under the collection's title
- **AND** named in words rather than as a file name

#### Scenario: a file the model says nothing about

- **WHEN** a file outside the model has changed
- **THEN** it SHALL still be listed, named from its own path
- **AND** SHALL NOT be left out

#### Scenario: publishing built pages

- **WHEN** a site that publishes built pages is published
- **THEN** the review afterwards SHALL say that nothing is waiting
- **AND** a page saved after that publish SHALL be the only change waiting
