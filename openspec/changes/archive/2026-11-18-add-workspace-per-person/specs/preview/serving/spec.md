# Spec Delta

## ADDED Requirements

### Requirement: A preview shows the asking person's own copy

A preview SHALL show the site as it is in the asking person's own copy, their
saves that have not yet reached the shared version included.

#### Scenario: two people, two changes

- **WHEN** two people each save a different change to the same page and both
  look at its preview
- **THEN** each SHALL see their own change in their preview
- **AND** neither SHALL see the other's change until it has reached their
  copy
