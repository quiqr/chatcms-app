# workspace/microvm Specification

## Purpose
Defines the micro VM workspace backend: what a VM can see and reach, how it
is started, found again and stopped.

## Requirements

### Requirement: A workspace can run in a micro VM of its own

An installation SHALL be able to run every workspace in a micro VM of its
own, chosen by configuration, with no change to any other behaviour of
workspaces or previews.

#### Scenario: the microvm backend

- **WHEN** the installation runs workspaces on the microvm backend and a
  person opens a project
- **THEN** their preview SHALL be served from a micro VM of their own

### Requirement: A VM sees only what its workspace needs

A VM SHALL see the host's package store read-only and the person's copy of
the site read-only, and nothing else of the host. It SHALL hold no key,
password or token of the installation.

#### Scenario: what is mounted

- **WHEN** a VM runs
- **THEN** its only shares SHALL be the package store and the person's copy,
  both read-only

### Requirement: A VM reaches nothing but the host

A VM SHALL have one network address of its own and SHALL reach no address
but the host's.

#### Scenario: the outside

- **WHEN** a VM tries to reach an address outside the host
- **THEN** it SHALL not get through

#### Scenario: another VM

- **WHEN** a VM tries to reach another VM
- **THEN** it SHALL not get through

### Requirement: Only as many VMs as there are addresses

A workspace SHALL be started only when an address is free for it; when none
is, starting SHALL be refused as the workspace limit is.

#### Scenario: no address free

- **WHEN** every address of the pool is taken
- **THEN** another workspace SHALL be refused with a reason worth showing

### Requirement: A VM that is stopped leaves nothing behind

Stopping a VM workspace SHALL end the VM and its shares and free its address.

#### Scenario: stopped

- **WHEN** a VM workspace is stopped
- **THEN** no process of it SHALL be left and its address SHALL be free again
