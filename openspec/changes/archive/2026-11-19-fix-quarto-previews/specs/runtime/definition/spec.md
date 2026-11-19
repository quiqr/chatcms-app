# Spec Delta

## MODIFIED Requirements

### Requirement: A preview that is not told its address is served at its root

A preview whose command has no `{url}` SHALL be addressed as
`<preview>/p/site/`, and SHALL receive each request with the preview's own
prefix and `p/site/` removed from the path. A preview told its address
SHALL receive the whole path.

#### Scenario: Quarto

- **WHEN** the preview command has no `{url}` and the browser asks for
  `<preview>/p/site/chapters/bread.html`
- **THEN** the preview SHALL be asked for `/chapters/bread.html`

#### Scenario: Quarto reloads itself

- **WHEN** a Quarto preview's source changes on disk
- **THEN** the page in the pane SHALL reload by itself and show the change

### Requirement: A preview may write beside the site

A preview SHALL be able to write files in the site's directory. In a micro
VM what it writes SHALL stay in the VM's memory and SHALL NOT reach the
person's copy. On the local backend a file it writes that git does not
track SHALL NOT count as unsaved work, and SHALL NOT keep other people's
changes from arriving.

#### Scenario: Quarto writes its output

- **WHEN** a Quarto preview in a micro VM renders
- **THEN** it SHALL succeed, and the person's copy SHALL hold no new file

#### Scenario: output in the copy on the local backend

- **WHEN** a preview has written `_book/` into the person's copy and
  another person's change arrives
- **THEN** the copy SHALL take the change, and SHALL not be shown as saving
