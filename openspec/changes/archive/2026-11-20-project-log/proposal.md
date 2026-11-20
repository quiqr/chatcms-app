# Proposal

Bean: [chatcms-qru3](../../../.beans/chatcms-qru3--log-in-preview-should-be-move-to-it-own-page.md)
(the preview's log moves to a page of its own).

## Why

The log of a preview start is shown inside the preview column and in the
workspace symbol's popover, is kept only in memory, and does not say which
Hugo version was asked for, which one was used and why, or on what backend
it ran. Debugging a project's Hugo version under gVisor needs exactly that,
after the fact. Builds and publications are recorded already but have no
place where an implementer reads them side by side.

## What Changes

- **Every preview start is kept** with its backend, who started it, how long
  it took, its steps, its log, and four facts about Hugo: the version the
  project asks for, the program the server chose and why (pinned through
  Nixhub, or the installed one because pinning is off or the project names
  no version), and the first line the program itself prints when it runs.
  Kept as long as the workspace history (`CHATCMS_WORKSPACE_HISTORY_DAYS`).
- **Project → Log**, in the Setup group, for whoever may see a project's
  technical detail: preview starts, builds and publications, newest first,
  filtered by kind; an entry opens to its detail and has its own address.
- **The preview column** says only that the preview is being prepared or did
  not start, with "Show the log" or "See what went wrong" for implementers;
  its folded log goes. The workspace symbol keeps the steps and links to the
  log instead of showing it. A failed publication links to its entry.

## Capabilities

### New Capabilities

- `history/project-log`: keeping preview starts and showing a project's
  log of starts, builds and publications.

### Modified Capabilities

- `workspace/startup`: the steps and log move out of the preview column.

## Impact

- `db/migrations`: `workspace_events` gains person, log and Hugo facts.
- `internal/workspace`: the start records them; `internal/app`: the Hugo
  resolver says why it chose a program.
- `internal/httpapi`: `GET …/log` and `GET …/log/{entry}`.
- `web/src`: the Log page, the Setup navigation, the preview column, the
  workspace symbol, the publish screen; strings.

## Assumptions

- Agreed on 20 November 2026: one page per area, starting with the project;
  git level, the assistant, organisations, the platform and access follow
  later (bean "Logs per area").
- "Log" sits under Setup, apart from History, which shows content versions
  to everyone.
- A start's log is capped (the last 400 lines) so one noisy start cannot
  fill the database.
- Technical words are allowed here: the page is for implementers.
