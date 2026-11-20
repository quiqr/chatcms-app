## ADDED Requirements

### Requirement: An answer the model cannot give ends with what went wrong

When the model refuses or keeps failing, the answer SHALL end, show what
went wrong in the person's chat, and stop the agent, rather than wait
without end: at once for a refusal, and after the second retry otherwise.

#### Scenario: the organisation's allowance is used up

- **WHEN** the gateway refuses the agent's request
- **THEN** the chat SHALL show the reason and stop working
- **AND** the agent's session SHALL be interrupted

#### Scenario: the model keeps failing

- **WHEN** the model fails and the agent has retried twice
- **THEN** the chat SHALL show the reason and stop working
