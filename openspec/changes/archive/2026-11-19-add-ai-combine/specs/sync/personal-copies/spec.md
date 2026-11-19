# Spec Delta

## ADDED Requirements

### Requirement: A copy with nothing of its own follows the shared version

A person's copy SHALL remember the version it last took from the shared
version. When it holds nothing written since, it SHALL take the shared
version as it is, even when the shared version's history was rewritten. A
copy with work of its own SHALL keep it.

#### Scenario: an answer replays waiting work

- **WHEN** an answer replays the shared version's waiting work on top of
  what arrived, and a person's copy holds nothing of its own
- **THEN** that copy SHALL show the answered version the next time it is
  opened
