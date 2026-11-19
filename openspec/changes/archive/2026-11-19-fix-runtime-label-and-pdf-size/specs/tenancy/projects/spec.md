# Spec Delta

## ADDED Requirements

### Requirement: A project says what it runs on

A project in the list SHALL say what its site runs on: the program its
runtime definition previews with, Hugo for a site without a definition, and
nothing for a project without a site.

#### Scenario: a Quarto book

- **WHEN** a project's site declares a Quarto runtime
- **THEN** its card SHALL say Quarto
