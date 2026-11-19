# Spec Delta

## ADDED Requirements

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
