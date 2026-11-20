# Design

## Decisions

- **Same events to the panel.** The relay translates OpenCode 2's events
  into the existing `text`, `tool`, `done` and `error` events, so nothing in
  the frontend changes. Text comes only from `session.text.delta`, which
  carries the assistant's text and never the person's own message.
- **Tool names** are remembered from `session.tool.input.started` by call
  id, because `session.tool.called` and `session.tool.failed` carry only the
  id; the `quiqr_` prefix is dropped as before.
- **Ending a stuck answer.** `session.retry.scheduled` carries the error and
  the attempt. A 4xx status, or an attempt of 3 or more, sends an `error`
  event with the gateway's message and interrupts the session; the
  `session.execution.interrupted` that follows ends the answer as `done`.
- **One nixpkgs.** `opencodeFor` takes OpenCode from the main nixpkgs; the
  `nixpkgs-opencode` input and its lock entry go.
