## MODIFIED Requirements

### Requirement: Three formats are read, chosen by extension

Model files SHALL be read as YAML, TOML or JSON according to their extension,
and a single site SHALL be able to mix formats across its model files. A TOML
array of tables (`[[build]]`, `[[collections]]`, `[[singles.fields]]`) SHALL
be read as a list, the same as a list in YAML or JSON.

#### Scenario: A TOML base with YAML includes

- **WHEN** a site has `quiqr/model/base.toml` and `quiqr/model/includes/collections.yaml`
- **THEN** both are read with the right parser
- **AND** the assembled model contains the collections from the include

#### Scenario: Tables in a list in TOML

- **WHEN** `quiqr/model/base.toml` holds `[[build]]` and `[[collections]]`
  entries, with `[[collections.fields]]` inside them
- **THEN** each is read as a list of entries
- **AND** the model is read without an "expected a list" error

#### Scenario: A file that is not valid in its format

- **WHEN** a model file cannot be parsed
- **THEN** the error names that file and the position of the problem
- **AND** no partially assembled model is returned
