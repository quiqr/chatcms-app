## ADDED Requirements

### Requirement: A model declares the shortcodes its pages may use

The model SHALL accept a `shortcodes` list in the base file and in
`includes/shortcodes/`, each entry with a `key`, a `title`, `fields` of the
single-value field types, and optionally `inner` set to `markdown` or `text`.
A field MAY carry `position`, a whole number from 0, naming the positional
parameter it is. Declarations SHALL be merged and checked like singles and
collections, and a model without `shortcodes` SHALL read as before.

#### Scenario: A figure and a notice

- **WHEN** the base declares `figure` with fields `src` (picture) and
  `caption` (string), and an include declares `notice` with
  `inner: markdown` and a field `type` at position 0
- **THEN** the model holds both declarations with their fields, in order

#### Scenario: A mistake in a declaration

- **WHEN** a shortcode declares a list field, or two fields at the same
  position, or `inner: html`
- **THEN** reading the model says what is wrong and in which file

#### Scenario: The old app's model

- **WHEN** a model has no `shortcodes` key
- **THEN** it reads exactly as before

### Requirement: The model may name the shortcode dialect

Shortcode declarations SHALL say nothing about syntax. The base MAY name the
dialect with `shortcodeDialect`, which SHALL be one of the dialects this
version has.

#### Scenario: Naming Hugo

- **WHEN** the base says `shortcodeDialect: hugo`
- **THEN** the model holds the dialect hugo

#### Scenario: An unknown dialect

- **WHEN** the base says `shortcodeDialect: textile`
- **THEN** reading the model says the dialect is unknown, and in which file
