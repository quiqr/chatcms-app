# Spec Delta

## MODIFIED Requirements

### Requirement: The browser follows the model

What is listed, and in what order, SHALL come from the site's Quiqr model and
its menu, as content/content-navigation describes. A collection's items SHALL
be found the way the Quiqr desktop app finds them, so that a site shows the
same items in both.

#### Scenario: pages and collections

- **WHEN** a site is browsed
- **THEN** the groups SHALL be the menu's sections, or "Pages" and
  "Collections" when the model has no menu

#### Scenario: a collection's items

- **WHEN** a collection under `content/` is listed
- **THEN** its items SHALL be the content files in its folder and in every
  folder below it
- **AND** SHALL be in the same order every time

#### Scenario: a collection that does not look below its folder

- **WHEN** a collection sets `includeSubdirs: false`
- **THEN** only the files directly in its folder SHALL be its items

#### Scenario: a page bundle

- **WHEN** a collection's folder holds `spring-sale/index.md`
- **THEN** that SHALL be one item, called after its folder, `spring-sale`
- **AND** the pictures beside it SHALL NOT be items

#### Scenario: a section's list page

- **WHEN** a collection sets `hideIndex: true`
- **THEN** `_index.md` and its other extensions SHALL NOT be items
- **AND** `index.md` of a page bundle SHALL still be one

#### Scenario: a collection outside content

- **WHEN** a collection's folder is not under `content/`
- **THEN** every data file below it SHALL be an item, by its path
- **AND** no folder SHALL be treated as a page bundle

#### Scenario: a hidden page or collection

- **WHEN** the model marks something hidden
- **THEN** it SHALL NOT be listed

#### Scenario: a site with no model

- **WHEN** a site has no model
- **THEN** the answer SHALL say so rather than show an empty site

#### Scenario: the count and the list agree

- **WHEN** a collection's count is shown beside its name
- **THEN** it SHALL be the number of items its list holds

## ADDED Requirements

### Requirement: An item is shown by its title and sorted as the model asks

An item SHALL be shown by the `title` in its front matter, falling back to a
readable form of its file or folder name. Items SHALL be sorted by the
front-matter field the collection names as its `sortkey`, ascending, and by
their path when it names none or an item lacks the field.

#### Scenario: an item with a title

- **WHEN** `2017-06-13-a-quick-note.md` has the title "A Quick Note"
- **THEN** it SHALL be listed as "A Quick Note"

#### Scenario: an item without one

- **WHEN** an item has no title in its front matter
- **THEN** it SHALL be listed by its file or folder name with hyphens as
  spaces

#### Scenario: a sort key

- **WHEN** a collection names `date` as its sortkey
- **THEN** its items SHALL be in the order of their dates
- **AND** an item without a date SHALL come after those with one, by path

### Requirement: An item is reached by a key that stays the same

Each item SHALL have a key made from its collection's key and its path in the
collection without the extension, with a page bundle named by its folder. The
key SHALL be what opening, renaming, copying and removing ask for.

#### Scenario: a flat file

- **WHEN** `content/post/hello.md` is in the collection `posts`
- **THEN** its key SHALL be `posts/hello`, as before this change

#### Scenario: a bundle in a subfolder

- **WHEN** `content/post/2024/spring-sale/index.md` is in the collection
  `posts`
- **THEN** its key SHALL be `posts/2024/spring-sale`
- **AND** opening that key SHALL open that file
