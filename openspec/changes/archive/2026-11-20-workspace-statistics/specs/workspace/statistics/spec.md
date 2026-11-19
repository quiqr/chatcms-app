## ADDED Requirements

### Requirement: Workspace history is recorded

The workspace manager SHALL record every start with its cold start time and
backend, every start refused because the limit was reached, and every failed
start with its reason, and SHALL record every five minutes how many
workspaces are starting and warm. History older than the configured number
of days, 90 by default, SHALL be removed.

#### Scenario: a start

- **WHEN** a workspace starts and its preview answers after 1.4 seconds
- **THEN** a start is recorded with 1400 milliseconds and its backend

#### Scenario: the limit

- **WHEN** a start is refused because the limit is reached
- **THEN** a refused start is recorded

#### Scenario: old history

- **WHEN** history is older than the configured days
- **THEN** the hourly cleanup removes it

### Requirement: The platform admin sees how workspaces are used

The platform admin area SHALL have a Workspaces screen showing now how many
workspaces are starting, warm and sleeping against the limit and on which
backend, per organisation and per project; for the past 7 days the starts,
the median and 90th percentile cold start time, the starts refused at the
limit, the failed starts by reason and the most warm at once per day; and
the disk held by the people's copies per project.

#### Scenario: looking

- **WHEN** a platform admin opens Workspaces
- **THEN** the screen shows the counts against the limit, the backend, the
  past 7 days and the disk per project, at 1440 px and at 390 px

### Requirement: Only a platform admin sees workspace statistics

The statistics and their screen SHALL be refused to anybody who is not a
platform admin.

#### Scenario: an organisation's owner

- **WHEN** somebody who administers an organisation, but not the platform,
  asks for the statistics
- **THEN** they are refused
