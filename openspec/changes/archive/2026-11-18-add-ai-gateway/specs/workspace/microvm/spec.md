# Spec Delta

## MODIFIED Requirements

### Requirement: A VM reaches nothing but the host

A VM SHALL have one network address of its own and SHALL reach no address
but the host's, and on the host only the AI gateway's port.

#### Scenario: the outside

- **WHEN** a VM tries to reach an address outside the host
- **THEN** it SHALL not get through

#### Scenario: another VM

- **WHEN** a VM tries to reach another VM
- **THEN** it SHALL not get through

#### Scenario: the gateway

- **WHEN** a VM connects to the AI gateway's port on its host address
- **THEN** it SHALL get through

#### Scenario: any other port of the host

- **WHEN** a VM connects to any other port of the host
- **THEN** it SHALL not get through
