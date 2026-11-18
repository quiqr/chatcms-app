# runtime/definition Specification

## Purpose
Defines what a site may declare in `quiqr/runtime.yaml` about how it is
previewed and built, and what happens when that declaration is wrong.

## Requirements

### Requirement: A site may declare its runtime

A site MAY hold `quiqr/runtime.yaml` naming packages with versions, a preview
command with its type and the page to open, and a build command with its
output. Its preview and its builds SHALL then run those commands with those
packages.

#### Scenario: a declared preview

- **WHEN** a site declares a preview command and its packages
- **THEN** its preview SHALL run that command with those packages, and
  nothing else

#### Scenario: a declared build

- **WHEN** a site declares a build command and output
- **THEN** a publish that builds SHALL run that command and publish that
  output

### Requirement: Without a declaration, a site runs as before

A site without `quiqr/runtime.yaml` SHALL be previewed and built by Hugo at
the project's recorded version, with the same commands as before.

#### Scenario: no file

- **WHEN** a site has no runtime definition
- **THEN** its preview and builds SHALL behave as before this capability

### Requirement: Commands are words, with placeholders filled in by the host

Commands SHALL be lists of words, run without a shell. The words `{host}`,
`{port}`, `{url}`, `{reload}`, `{source}`, `{output}` and `{cache}` SHALL be
replaced by the workspace or the build; every other word SHALL be passed as
written. A command's program SHALL be one of the declared packages' programs.

#### Scenario: a shell trick

- **WHEN** a command holds `"hugo; rm -rf /"` as one word
- **THEN** it SHALL be passed to the program as that one word, not run

#### Scenario: a program from elsewhere

- **WHEN** a command names a program none of the declared packages provides
- **THEN** the preview or build SHALL be refused, naming the program

### Requirement: A wrong declaration is said, not guessed around

A runtime definition that cannot be read, names an unknown preview type or
placeholder, or holds no preview command SHALL refuse the preview with a
reason that names what is wrong, shown to implementers in the startup log.

#### Scenario: an unknown type

- **WHEN** the preview type is `video`
- **THEN** the preview SHALL be refused, saying the type is not one of web,
  pdf or other

### Requirement: A changed declaration takes effect when the preview is next opened

When the runtime definition in a person's copy differs from the one their
running preview started with, opening the preview SHALL start it afresh with
the new definition.

#### Scenario: the definition arrives from elsewhere

- **WHEN** a runtime definition reaches a person's copy while their preview
  runs
- **THEN** the next time they open the preview it SHALL run with the new
  definition
