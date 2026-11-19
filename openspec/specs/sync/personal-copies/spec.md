# sync/personal-copies Specification

## Purpose
Defines each person's own copy of a site: how it is made, what reads and
writes it, how saves made there reach the shared version, and what the shared
version still does alone.

## Requirements

### Requirement: Each person edits their own copy

Every person working on a project SHALL have their own copy of the site, made
from the shared version the first time they open it. Reading and saving
content, media and history SHALL use the asking person's copy.

#### Scenario: a first open

- **WHEN** a person opens a project for the first time
- **THEN** a copy of the site SHALL be made for them from the shared version

#### Scenario: an unfinished change

- **WHEN** one person saves a page and another person reads it before the
  save has reached the shared version
- **THEN** the second person SHALL see the page as it was

### Requirement: A save reaches the shared version by itself

After a save, the person's copy SHALL be brought level with the shared
version and then sent to it, without the person doing anything. Their own
versions SHALL be replayed on top of the shared version's, so history stays a
list of edits.

#### Scenario: nobody else changed anything

- **WHEN** a person saves and nobody else has changed the shared version
- **THEN** the save SHALL reach the shared version

#### Scenario: someone else changed another page

- **WHEN** a person saves while another person's save to a different page
  has reached the shared version
- **THEN** both saves SHALL be in the shared version
- **AND** the person's copy SHALL hold both

### Requirement: Two people changing the same thing are asked, not overwritten

When a person's copy and the shared version changed the same file, the
exchange SHALL stop and the question SHALL be put the way the settling
capability puts it. Until it is settled, the person's change SHALL stay in
their copy and SHALL NOT reach the shared version.

#### Scenario: the same field

- **WHEN** two people change the same field of the same page
- **THEN** the second exchange SHALL stop
- **AND** the second person SHALL be asked about that field
- **AND** the first person's change SHALL be in the shared version untouched

### Requirement: The shared version alone talks to the outside

Exchanging with the site's remote and publishing SHALL use the shared version
only. A person's copy SHALL NOT reach the remote, and SHALL NOT need the
project's key.

#### Scenario: publishing

- **WHEN** somebody publishes
- **THEN** what is published SHALL be the shared version
- **AND** the publishing person's own saves SHALL have reached it first

#### Scenario: a person's copy

- **WHEN** a person's copy is brought level or sent
- **THEN** it SHALL exchange with the shared version on this machine only

### Requirement: What arrives from the remote reaches every copy

Versions that arrive in the shared version from the remote SHALL reach a
person's copy the next time it is opened or saved.

#### Scenario: a change made outside ChatCMS

- **WHEN** someone changes the site's remote outside ChatCMS and the shared
  version takes the change in
- **THEN** a person's copy SHALL hold that change the next time it is opened

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
