# Spec Delta

## ADDED Requirements

### Requirement: A site kept by ChatCMS can be connected later

Somebody allowed to change a project's settings SHALL be able to connect a
site kept by ChatCMS to a repository. The project's key SHALL reach it and it
SHALL be empty. Every version so far SHALL then be sent there, the address
recorded, and the copy ChatCMS kept removed.

#### Scenario: Connecting to an empty repository

- **WHEN** a kept site with three versions is connected to an empty
  repository the key reaches
- **THEN** that repository SHALL hold the same three versions
- **AND** the project SHALL no longer say it is kept by ChatCMS
- **AND** the copy ChatCMS kept SHALL be gone

#### Scenario: A repository that holds work

- **WHEN** a kept site is connected to a repository that is not empty
- **THEN** connecting SHALL be refused with a message saying so
- **AND** nothing SHALL be sent and the site SHALL stay kept by ChatCMS

#### Scenario: A repository the key does not reach

- **WHEN** a kept site is connected to a repository that refuses the key
- **THEN** connecting SHALL be refused with a message saying what to do
- **AND** the site SHALL stay kept by ChatCMS

#### Scenario: A site that is already connected

- **WHEN** connecting is asked for on a site that is not kept by ChatCMS
- **THEN** it SHALL be refused

#### Scenario: An editor

- **WHEN** somebody who may only edit content asks to connect the site
- **THEN** it SHALL be refused by the permission

### Requirement: Removing a project removes what ChatCMS kept for it

Removing a project whose site is kept by ChatCMS SHALL remove the copy
ChatCMS kept, since nothing else holds it.

#### Scenario: Removing a kept project

- **WHEN** a project whose site is kept by ChatCMS is removed
- **THEN** its kept repository SHALL be gone
