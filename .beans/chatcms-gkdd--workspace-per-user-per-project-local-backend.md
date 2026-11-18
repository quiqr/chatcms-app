---
# chatcms-gkdd
title: Workspace per user per project, local backend
status: in-progress
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T13:09:24Z
parent: chatcms-z0p7
---

## Purpose

The backend asks for workspace(user, project) and gets back an address and a
lifecycle. Where the workspace runs is hidden behind the interface. The first
backend runs local processes, as the preview does today.

## Scope

- `internal/workspace`: the interface and a local backend (processes,
  bubblewrap). `internal/preview` moves behind it.
- Keyed on (user, project) instead of project: each user gets their own
  working copy on their own branch.
- Lifecycle owned by ChatCMS (start, warm, sleep after idle, cleanup), state
  in Postgres, reconciled on start-up like `InterruptLeftovers`.
- Workspace limits replace `PreviewsAtOnce`; the idle sweep carries over.
- Eager start when a user opens a project, with the startup UX from the
  briefing.
- Builds stay one-shot jobs through `runner.Runner`.

## Done when

- Two users preview the same project at the same time with different changes,
  asserted in e2e at both widths.
- A server restart reattaches to or cleans up running workspaces, proven by an
  integration test.
- A fake backend in the tests shows that a backend swap needs no change
  outside `internal/workspace`.

## Open decisions

- `gitsync` speaks in sites and versions and takes no branch names. A per-user
  working copy is added in that vocabulary ("this person's copy of the site"),
  not by exposing branches. Flag the naming in the proposal.

## References

`docs/BRIEFING-M07.md` step 1 and the addendum on backends, lifecycle and
startup UX. Builds on chatcms-ihey (live preview).
