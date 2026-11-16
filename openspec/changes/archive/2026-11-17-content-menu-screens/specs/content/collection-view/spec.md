# Spec Delta

## ADDED Requirements

### Requirement: A collection is shown as a list of its items

Picking a collection SHALL show its items in a list headed by the collection's
title and a line such as "48 posts · newest first", using the collection's
own word for one item. Each row SHALL show the item's title and, where the
model has them, its date, author and folder, a thumbnail from the first image
field, and a status badge.

#### Scenario: A row

- **WHEN** a post with a title, a date, an author and a picture is listed
- **THEN** its row shows the title, the date, the author and a thumbnail
- **AND** the title opens the item

#### Scenario: Status

- **WHEN** items are changed, new, set not to show on the website, or
  published
- **THEN** their badges say Changed, New, Draft and Published
- **AND** on a phone the Published badge is left out

#### Scenario: Paging

- **WHEN** a collection holds 48 items
- **THEN** 20 are shown with "Showing 20 of 48" and a button for 20 more

#### Scenario: An empty collection

- **WHEN** a collection holds no items
- **THEN** it says there are none yet and offers to make one, when the person
  may

### Requirement: A collection can be searched, filtered by folder and sorted

The list SHALL offer a search over titles and summaries, a choice of folder
when the collection has subfolders, and a sort. The server SHALL apply them,
so that the count and paging are of the filtered list.

#### Scenario: Searching

- **WHEN** somebody types "spring" in the collection's search
- **THEN** after a short pause only items whose title or summary holds it are
  listed, with their number

#### Scenario: No results

- **WHEN** nothing matches
- **THEN** the list says no items match, in the collection's own word, and
  offers to clear the search

#### Scenario: Folders

- **WHEN** a collection has the subfolders `nl` and `en`
- **THEN** a choice of "All", "nl" and "en" is offered, each with its count
- **AND** with more than five folders the choice is a list rather than tabs
- **AND** without subfolders no choice is shown

#### Scenario: Sorting

- **WHEN** somebody chooses oldest first, title, or recently edited
- **THEN** the list is in that order
- **AND** the default is the model's sort key, else newest first

### Requirement: File names are for those who see technical detail

Those the policy allows to see technical detail SHALL be able to switch on
each item's file name in the list. Nobody else SHALL be sent a path.

#### Scenario: An implementer

- **WHEN** an implementer switches file names on
- **THEN** each row shows its file under the title

#### Scenario: An editor

- **WHEN** an editor lists a collection
- **THEN** no switch is offered and the answer holds no paths

### Requirement: Making an item checks its address first

The new-item dialog SHALL ask for a title, and a folder when the collection
has subfolders. It SHALL show the web address made from the title, which can
be changed. For those who see technical detail it SHALL show the file that
will be made. A taken address SHALL be refused before anything is written,
and the dialog SHALL say how to fix it.

#### Scenario: Making a post

- **WHEN** somebody types "Spring sale" and chooses Create and edit
- **THEN** the item is made and opens in the editor

#### Scenario: A taken address

- **WHEN** the address made from the title is already used
- **THEN** the dialog says there is already a post at that address and to
  change the title or the address
- **AND** Create and edit stays unavailable

#### Scenario: A viewer

- **WHEN** a viewer lists a collection
- **THEN** no new, rename, duplicate or remove action is offered

### Requirement: Each row offers what can be done to it

Each row SHALL offer Open, and for those who may edit, Rename, Duplicate and
Remove, following content/pages.

#### Scenario: Removing from the list

- **WHEN** somebody removes an item from its row
- **THEN** they are told first that it goes at the next publish and can be
  brought back
- **AND** after confirming, the list no longer holds it
