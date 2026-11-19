# Design

## Context

`GET …/workspace` answers a person's workspace standing: state, reason, and
for implementers the steps and the log. The manager knows the backend
(`Workspaces.Backend()`) and when a start began (`live.began`), and can put
a workspace to sleep (`Sleep`). The assistant answers its own availability
at `GET …/chat`.

## Decisions

- **One component, `WorkspaceIndicator`**, beside `SyncIndicator` in every
  shell that shows it. It reads the standing and the assistant itself, so a
  shell only places it.
- **Restart is sleep then open**, on the server, so the client needs one
  call and the person's copy is kept. It answers the new standing.
- **Technical detail stays behind `ActionSeeTechnicalDetail`**: the backend
  name, steps, log and restart. The restart endpoint refuses others with 403.
- **Since** is sent as RFC 3339 and shown as "Started" with the app's one
  way of writing a moment (`when`): minutes ago, a time, or a day.
- **The preview column follows a restart**: the symbol announces it with a
  window event, and the preview opens the workspace again.

## Risks

- A restart while the assistant works ends its session; the button says the
  preview starts again, and asks nothing more, because only implementers see
  it.
