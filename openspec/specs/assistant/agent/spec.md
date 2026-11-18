# assistant/agent Specification

## Purpose
Defines the agent that runs in a workspace: what starts it, what it may reach,
and what it is told.

## Requirements

### Requirement: Each workspace runs its own agent

When the AI gateway is configured, a workspace SHALL run an agent beside its
preview, working on the person's copy of the site. Without a gateway or an
agent program, the workspace SHALL run without one and the preview SHALL work
as before.

#### Scenario: a gateway and an agent

- **WHEN** a workspace starts on an installation with a gateway and an agent
- **THEN** an agent SHALL run in it, reachable by the host

#### Scenario: no gateway

- **WHEN** no gateway is configured
- **THEN** the workspace SHALL start with no agent, and its preview SHALL work

### Requirement: The agent reaches the model only through the gateway

The agent SHALL be configured with the gateway as its only model provider and
the workspace's token as its key. It SHALL fetch nothing from the internet.

#### Scenario: its configuration

- **WHEN** the agent starts
- **THEN** its only provider SHALL be the gateway, with the workspace's token

### Requirement: The site cannot reconfigure the agent

The agent's configuration SHALL come from the host alone; configuration files
in the site SHALL be ignored.

#### Scenario: a site with its own agent configuration

- **WHEN** a site holds a configuration file for the agent that allows more
- **THEN** the agent SHALL keep the host's configuration

### Requirement: Only the host may talk to the agent

The agent's API SHALL require a secret only the host holds.

#### Scenario: without the secret

- **WHEN** anything asks the agent's API without the secret
- **THEN** it SHALL be refused
