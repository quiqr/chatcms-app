# onboarding/import Specification

## Purpose
Defines how a site arrives in ChatCMS, what is checked about it, and what an
implementer is told, so that importing a real customer's site is a few minutes
rather than an afternoon.

## Requirements

### Requirement: A project has a key of its own

Each project SHALL have its own key pair for reaching its remote. The public
half SHALL be available to show; the private half SHALL be stored encrypted.

#### Scenario: A key is made with the project

- **WHEN** a project is set up
- **THEN** it has a key pair, and the public half can be read and copied

#### Scenario: Two projects do not share a key

- **WHEN** two projects are set up
- **THEN** their keys differ

#### Scenario: Replacing a key

- **WHEN** an implementer asks for a new key
- **THEN** a new pair replaces the old one
- **AND** the new public half is shown, with a reminder that the git host
  needs it

### Requirement: A site arrives from a git repository, an export, or a template

Importing SHALL accept a git repository address, a directory holding a site
exported from the Quiqr desktop app, or a template from the template list. An
export or a template MAY come without an address, and is then kept by ChatCMS
until it is connected.

#### Scenario: From a repository

- **WHEN** a repository address is given
- **THEN** the site is fetched into the project

#### Scenario: From an export

- **WHEN** a directory holding an exported site is given
- **THEN** the site becomes the project's content
- **AND** it is sent to the project's remote

#### Scenario: From a template

- **WHEN** a template is chosen and the project's remote is an empty
  repository the deploy key can reach
- **THEN** the site starts as that template's files in one version named
  after the template, with none of the template's history
- **AND** it is sent to the project's remote

#### Scenario: From a template into a repository that is not empty

- **WHEN** a template is chosen and the project's remote already holds work
- **THEN** importing SHALL be refused with a message saying the repository
  is not empty
- **AND** nothing SHALL be sent to it

#### Scenario: From a template that is not in the list

- **WHEN** the chosen template is not in the template list
- **THEN** importing SHALL be refused

#### Scenario: An export and its repository agree

- **WHEN** the same site is imported once as an export and once as the
  repository made from that export
- **THEN** the two projects hold the same content

#### Scenario: A template without an address

- **WHEN** a template is chosen and no address is given
- **THEN** the site starts as that template's files in one version named
  after the template
- **AND** the project SHALL say its site is kept by ChatCMS and not connected
- **AND** no key SHALL be needed

#### Scenario: An export without an address

- **WHEN** a directory holding an exported site is given and no address
- **THEN** the site becomes the project's content
- **AND** the project SHALL say its site is kept by ChatCMS and not connected

#### Scenario: A repository import without an address

- **WHEN** a git import is asked for with no address
- **THEN** importing SHALL be refused, because the address is where the site
  comes from

### Requirement: What arrived is inspected and reported plainly

After importing, the system SHALL report whether the site is a Hugo site and
whether it carries a Quiqr model, in words an implementer can act on.

#### Scenario: A Hugo site with a model

- **WHEN** the imported site has a Hugo configuration and a Quiqr model
- **THEN** both are reported as found

#### Scenario: A Hugo site with no model

- **WHEN** the imported site has a Hugo configuration and no Quiqr model
- **THEN** it is reported as a Hugo site that still needs a model
- **AND** the project is usable rather than broken

#### Scenario: Something that is not a Hugo site

- **WHEN** the imported site has no Hugo configuration
- **THEN** that is said plainly, naming what was looked for

#### Scenario: A model that does not parse

- **WHEN** the imported site has a model that cannot be read
- **THEN** the reason is reported, naming the file and what was expected

### Requirement: The connection can be tested on demand

An implementer SHALL be able to ask whether the project's key reaches its
remote, and be told what to do rather than what failed.

#### Scenario: The key works

- **WHEN** the remote accepts the key
- **THEN** the answer says the connection is good

#### Scenario: The key has not been added yet

- **WHEN** the remote is reachable and refuses the key
- **THEN** the answer says the key still has to be added to the git host, and
  shows it again

#### Scenario: The remote cannot be reached

- **WHEN** the remote does not answer at all
- **THEN** the answer says so, and distinguishes it from a refused key

#### Scenario: Nothing in the answer is secret

- **WHEN** any of these answers is given
- **THEN** it carries no private key and no credential
