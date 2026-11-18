# Spec Delta

## ADDED Requirements

### Requirement: A VM runs the agent for the host only

A VM SHALL run the workspace's agent when it was given a token, and the
agent's port SHALL be reached from the host only, behind the agent's secret.

#### Scenario: an agent in a VM

- **WHEN** a VM starts with a token
- **THEN** the agent SHALL answer the host on its port, with the secret
