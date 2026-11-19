# Spec Delta

## Purpose

Defines how a workflow names what its build produces and where each part
goes.

## ADDED Requirements

### Requirement: A workflow may list its build's outputs

A workflow that builds MAY list outputs, each a path inside the build's
output and a destination: `publish`, `artifact` or `both`. A path is a
file, a directory, or a pattern; `.` is everything. A workflow that lists no
outputs SHALL publish the whole build, as before.

#### Scenario: a book

- **WHEN** a publish workflow lists `.` to publish and `book.pdf` to
  artifact
- **THEN** the site SHALL be published and the PDF SHALL be kept

#### Scenario: outputs without a build

- **WHEN** a workflow lists outputs and has no build step
- **THEN** reading SHALL fail, saying outputs come from a build

#### Scenario: something to publish but no publishing step

- **WHEN** an output goes to publish and the workflow does not publish
- **THEN** reading SHALL fail, naming the output

#### Scenario: an unknown destination

- **WHEN** an output's destination is not publish, artifact or both
- **THEN** reading SHALL fail, naming the destinations

### Requirement: Publishing takes only what goes to publish

When outputs are listed, the publishing step SHALL be given only the files
of outputs that go to publish, and the outputs that go to artifact SHALL be
handed to the artifact store once every step has succeeded.

#### Scenario: an output only for artifact

- **WHEN** `book.pdf` goes to artifact only
- **THEN** the published site SHALL NOT hold it

#### Scenario: a failed step

- **WHEN** a step fails
- **THEN** no artifact SHALL be kept
