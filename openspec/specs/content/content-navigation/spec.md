# content/content-navigation Specification

## Purpose
Defines the navigation structure a site's model implies: the pages and
collections a content creator browses, in the order and under the labels the
model asks for.

## Requirements

### Requirement: Navigation is derived from the model

The system SHALL derive the content navigation from the assembled model, not
from a second structure maintained by hand. A change to the model SHALL change
the navigation with no further step.

#### Scenario: Pages and collections appear

- **WHEN** a model declares two singles and three collections
- **THEN** the navigation offers both groups, with those entries

#### Scenario: A collection reports how many items it has

- **WHEN** a collection's folder holds seven content files
- **THEN** the navigation reports seven for that collection

#### Scenario: An empty collection still appears

- **WHEN** a collection's folder is empty or absent
- **THEN** the collection still appears, reporting none

### Requirement: Labels come from the model

Each collection SHALL carry the singular item name the model gives it, so that
the interface can say "New article" rather than "New item".

#### Scenario: A collection with an item title

- **WHEN** a collection declares its item name
- **THEN** the navigation carries that name for use in labels

#### Scenario: A collection without one

- **WHEN** a collection declares no item name
- **THEN** the navigation carries a neutral fallback
- **AND** the fallback is not a git term or a developer term

### Requirement: The model's menu decides the order

When a model declares a menu, the navigation SHALL follow its sections and
order, and the content screen SHALL show those sections. When it does not,
there SHALL be a group called "Pages" holding the singles and a group called
"Collections" holding the collections, each in the order the model declares.

#### Scenario: A declared menu

- **WHEN** a model declares a menu with two sections
- **THEN** the navigation has those sections, in that order, with the entries
  each lists
- **AND** the content screen shows the same sections in the same order

#### Scenario: A menu naming a single and a collection

- **WHEN** one menu section names a single and a collection
- **THEN** the single is an entry that opens its form
- **AND** the collection is an entry that carries how many items it has

#### Scenario: A menu naming something that is not there

- **WHEN** a menu names a single or collection the model does not declare
- **THEN** the entry is omitted from the navigation
- **AND** the omission is reported as a model warning, not as a failure

#### Scenario: No menu declared

- **WHEN** a model declares no menu
- **THEN** there is a "Pages" group with the singles, then a "Collections"
  group with the collections, each in declaration order

#### Scenario: A single and a collection share a key

- **WHEN** a model declares a single and a collection with the same key
- **THEN** a model warning says so
- **AND** each still opens what it names

### Requirement: Hidden entries stay hidden

An entry the model marks as hidden SHALL be absent from the navigation while
remaining reachable by its key.

#### Scenario: A hidden collection

- **WHEN** a collection is marked hidden
- **THEN** it does not appear in the navigation
- **AND** asking for it by key still returns it

### Requirement: A developer's section is shown only to those who see technical detail

A menu section with `matchRole: siteDeveloper` SHALL be shown only to people
the policy allows to see technical detail. A section marked `disabled` SHALL
NOT be shown to anybody. Any other `matchRole` SHALL NOT hide a section.

#### Scenario: An editor

- **WHEN** an editor opens the content of a site whose menu has a section for
  site developers
- **THEN** that section is not shown
- **AND** its singles stay reachable by key for those allowed to edit them

#### Scenario: An implementer

- **WHEN** an implementer opens the same site
- **THEN** that section is shown, in its place in the menu

#### Scenario: A disabled section

- **WHEN** a menu section is marked disabled
- **THEN** nobody is shown it

#### Scenario: Another role name

- **WHEN** a section names `matchRole: contentEditor`
- **THEN** everybody who may read the site is shown it
