## MODIFIED Requirements

### Requirement: A format that cannot be written faithfully is refused

Front matter in a format the system cannot write the way the reference writes
it SHALL be readable, and writing it SHALL fail with a message naming the
format, rather than producing a file that differs from what the reference
would have produced. This applies to front matter only; a data file is
written whole (see "A data file is read and written whole").

#### Scenario: Reading a format that cannot be written

- **WHEN** a content file uses front matter in such a format
- **THEN** its values and body are readable

#### Scenario: Writing it is refused

- **WHEN** such a document is written
- **THEN** writing fails with a message naming the format
- **AND** no file is produced

#### Scenario: The format every site uses is writable

- **WHEN** a document whose front matter is YAML is written
- **THEN** it succeeds

## ADDED Requirements

### Requirement: A data file is read and written whole

A file whose extension is not `.md`, `.markdown`, `.html` or `.qmd` SHALL be
read whole as JSON, YAML or TOML according to its extension, and written
whole in the same format, matching the reference: JSON as
`JSON.stringify(value, null, "  ")` with no final newline, YAML as js-yaml
writes it, TOML as tomlify-j0.4 with `space: 2` with no final newline. The
order of keys at every depth SHALL be kept.

#### Scenario: A TOML configuration

- **WHEN** `config.toml` holds `title = "Old"` and a `[params]` table
- **THEN** the form shows the title and the params
- **AND** saving a new title writes TOML with no front matter delimiters and
  the params in their order

#### Scenario: YAML data

- **WHEN** `data/hours.yaml` holds `note: Closed on Sundays`
- **THEN** the form shows "Closed on Sundays"
- **AND** saving writes YAML without `---`

#### Scenario: JSON data

- **WHEN** `data/services.json` is edited and saved
- **THEN** it is written as the reference writes it, with two spaces of
  indent and no final newline

#### Scenario: The reference's bytes

- **WHEN** each recorded case is written as JSON and as TOML
- **THEN** the bytes equal what the reference produced
