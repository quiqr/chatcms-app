# ai/gateway Specification

## Purpose
Defines who may call a language model through the host, with what, and what
the host passes on to the provider.

## Requirements

### Requirement: Only a running workspace's token is accepted

The gateway SHALL accept a request only with the token of a workspace that is
starting or warm. A token SHALL stop working when its workspace sleeps, fails
to start or is started again.

#### Scenario: a running workspace

- **WHEN** a request carries the token of a warm workspace
- **THEN** it SHALL be passed on to the provider

#### Scenario: no token, or a stopped workspace's

- **WHEN** a request carries no token, an unknown one, or the token of a
  workspace that has gone to sleep
- **THEN** it SHALL be refused as unauthorised
- **AND** nothing SHALL be sent to the provider

### Requirement: The provider's key never leaves the host

The gateway SHALL send the provider's key in place of the workspace's token
and SHALL never return the key, in an answer or an error.

#### Scenario: passing a request on

- **WHEN** a request is passed on
- **THEN** the provider SHALL receive the installation's key
- **AND** the workspace's token SHALL not be sent to the provider

### Requirement: Only the configured model is used

The gateway SHALL answer only the chat completion and model list routes, and
SHALL use the configured model whatever model a request names.

#### Scenario: another model asked for

- **WHEN** a request names a different model
- **THEN** the configured model SHALL be used

#### Scenario: another route

- **WHEN** a request asks for any other route
- **THEN** it SHALL be answered as not found

### Requirement: Streamed answers pass through as they come

A streamed answer SHALL reach the workspace piece by piece as the provider
sends it.

#### Scenario: a streamed answer

- **WHEN** a workspace asks for a streamed answer
- **THEN** each piece SHALL be passed on as it arrives

### Requirement: The gateway is reached only from workspaces

The gateway SHALL listen apart from the app, and SHALL answer only requests
from this machine or from a workspace's address.

#### Scenario: from elsewhere

- **WHEN** a request comes from any other address
- **THEN** it SHALL be refused
