# onboarding/template-gallery Specification

## Purpose
Defines how ChatCMS learns which templates exist, how it shows them, and how
somebody chooses one to start a project from, so that a new site can begin
from a Quiqr community template.

## Requirements

### Requirement: The server keeps a validated copy of the template feed

The server SHALL fetch the configured template feed in the background, keep
every entry that has a name, a normalised name and a source repository, and
drop the rest. On a failed fetch it SHALL keep serving the last good copy.

#### Scenario: A feed with one broken entry

- **WHEN** the feed holds nine entries and one has no source repository
- **THEN** the template list SHALL hold the other eight

#### Scenario: The feed cannot be reached after a good fetch

- **WHEN** a refresh fails after an earlier one succeeded
- **THEN** the template list SHALL still be the earlier one

#### Scenario: The feed was never reached

- **WHEN** no fetch has ever succeeded
- **THEN** asking for the template list SHALL say the gallery is unavailable
  rather than return an empty list

### Requirement: Templates are listed alphabetically with their details

The template list SHALL give, per template, its name, description, author and
author link, licence and licence link, homepage, live demo address, Hugo
theme, Hugo version and whether it is extended, content model parts, number
of forms, source repository and a screenshot address served by ChatCMS. It
SHALL be sorted by name.

#### Scenario: Feed order differs

- **WHEN** the feed lists Summer before Paper
- **THEN** the list SHALL give Paper before Summer

#### Scenario: Hugo version from the feed

- **WHEN** an entry's Hugo version is `extended_0.101.0`
- **THEN** the list SHALL give version `0.101.0` with extended set

### Requirement: Screenshots are served by ChatCMS

A template's screenshot SHALL be fetched by the server, cached, and served
from ChatCMS's own address, so that no browser contacts the feed's host to
show the gallery.

#### Scenario: Showing the gallery

- **WHEN** the gallery is shown
- **THEN** every screenshot address SHALL be one of ChatCMS's own

#### Scenario: A screenshot that cannot be fetched

- **WHEN** a template's screenshot cannot be fetched
- **THEN** the card SHALL show a placeholder with the template's name

### Requirement: The gallery shows, searches and explains

Starting a project from a template SHALL show the templates as cards with
screenshot, name, description, author and licence, a live demo link when the
template has one, and a button to use it. A search SHALL filter by name,
description and author, ignoring case.

#### Scenario: Searching

- **WHEN** somebody types `recipe` into the search
- **THEN** only templates whose name, description or author holds it SHALL
  be shown
- **AND** the count SHALL say how many of how many are shown

#### Scenario: Nothing matches

- **WHEN** the search matches no template
- **THEN** the gallery SHALL say no template matches what was typed

#### Scenario: The gallery is unavailable

- **WHEN** the template list is unavailable
- **THEN** the gallery SHALL say so, offer to try again, and point to
  importing a git repository instead

#### Scenario: A template without a demo

- **WHEN** a template has no live demo address
- **THEN** its card SHALL show no live demo link

#### Scenario: On a phone

- **WHEN** the gallery is shown at 390 px wide
- **THEN** the cards SHALL stand in one column with no sideways scrolling

### Requirement: Choosing a template asks for a name and an empty repository

Using a template SHALL lead to a step that shows the template's details and
asks for a project name, the organisation and the address of an empty
repository, before the deploy key step.

#### Scenario: A name already used

- **WHEN** the name is taken by a project in that organisation
- **THEN** continuing SHALL be refused with a message saying so

#### Scenario: The last free project

- **WHEN** the organisation has one project left under its limit
- **THEN** the step SHALL say this uses the last one

#### Scenario: No free project

- **WHEN** the organisation is at its project limit
- **THEN** continuing SHALL be refused with the limit's own message

#### Scenario: A different Hugo version

- **WHEN** the template asks for a Hugo version other than the one this
  installation runs
- **THEN** the step SHALL say which version the template asks for and which
  one builds it here

### Requirement: Only those who may create projects see the gallery

The gallery and the step after it SHALL be offered only to people who may
create projects in the organisation.

#### Scenario: An editor

- **WHEN** somebody who may only edit content opens the address of the
  gallery
- **THEN** it SHALL be refused by the permission rather than by a hidden link
