# workspace/gvisor Specification

## Purpose
TBD - created by archiving change add-gvisor-workspaces. Update Purpose after archive.

## Requirements

### Requirement: A workspace can run in a gVisor sandbox of its own

An installation SHALL be able to run every workspace in its own gVisor
sandbox, chosen by configuration alone, without root privileges and without
hardware virtualisation.

#### Scenario: choosing gVisor

- **WHEN** an installation sets its workspace backend to gvisor
- **THEN** opening a project starts the person's preview and agent in a
  sandbox of their own, and the preview shows the site

### Requirement: A sandbox sees only what its workspace needs

A sandbox SHALL see the Nix store read-only, the person's copy of the site,
a private temporary directory and its socket directory, and nothing else of
the host's files.

#### Scenario: looking around

- **WHEN** the preview command lists the root of its file system
- **THEN** it finds only those directories

### Requirement: A sandbox reaches nothing but its sockets

A sandbox SHALL have no network. The preview and the agent SHALL be reached
from the host, and the AI gateway from inside, only through the workspace's
Unix sockets.

#### Scenario: no way out

- **WHEN** something inside the sandbox tries to reach an address on the
  internet or on the host's network
- **THEN** it fails

#### Scenario: the way in

- **WHEN** the proxy asks the workspace for the preview
- **THEN** the request reaches the preview server inside through the preview
  socket

### Requirement: A stopped sandbox leaves nothing behind

Stopping a workspace SHALL end its sandbox and its relays and remove its
bundle, state and sockets.

#### Scenario: stopping

- **WHEN** a workspace is stopped
- **THEN** no process, bundle, state directory or socket of it remains
