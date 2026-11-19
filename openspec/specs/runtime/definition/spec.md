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

### Requirement: A package may name the nixpkgs revision it comes from

A package in the runtime definition MAY be given as a version and a nixpkgs
revision. It SHALL then be built from that revision without asking the
package index, and refused when that revision holds another version.

#### Scenario: a pinned revision

- **WHEN** a site declares `quarto: { version: "1.10.18", nixpkgs: <rev> }`
- **THEN** Quarto SHALL be built from that revision

#### Scenario: the revision holds another version

- **WHEN** the named revision's package has another version
- **THEN** the preview SHALL be refused, saying which version it holds

### Requirement: An installation may provide packages itself

An installation MAY name packages it provides, each with a store path. A
declared package SHALL be taken from there when the store path is of the
version asked for, before any is built.

#### Scenario: a provided package

- **WHEN** the installation provides `quarto` at a store path of 1.10.18 and
  a site asks for 1.10.18
- **THEN** that store path SHALL be used

#### Scenario: another version

- **WHEN** a site asks for a version the provided path is not
- **THEN** the provided path SHALL NOT be used

### Requirement: A preview that is not told its address is served at its root

A preview whose command has no `{url}` SHALL receive each request with the
preview's own prefix removed from the path. A preview told its address SHALL
receive the whole path.

#### Scenario: Quarto

- **WHEN** the preview command has no `{url}` and the browser asks for
  `<preview>/chapters/bread.html`
- **THEN** the preview SHALL be asked for `/chapters/bread.html`

### Requirement: A preview may write beside the site

A preview SHALL be able to write files in the site's directory. In a micro
VM what it writes SHALL stay in the VM's memory and SHALL NOT reach the
person's copy.

#### Scenario: Quarto writes its output

- **WHEN** a Quarto preview in a micro VM renders
- **THEN** it SHALL succeed, and the person's copy SHALL hold no new file
