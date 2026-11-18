---
# chatcms-gkdd
title: Workspace per user per project, local backend
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T14:44:24Z
openspec-link: openspec/changes/archive/2026-11-18-add-workspace-per-person
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-18-add-workspace-per-person`.

- `internal/workspace`: a workspace per (person, project), state in
  Postgres, cap taken in one transaction, idle sweep, hourly cleanup,
  reconcile on start-up, startup steps and log. The `local` backend runs
  Hugo; its workspaces stop with the app, a backend that runs apart keeps
  them across a restart.
- A copy of the site per person, cloned from the shared copy, which is its
  remote. Saves reach the shared version through a `site-share` job using
  the existing exchange and settling; a person waiting on a question keeps
  seeing their own work. Only the shared copy reaches the remote.
- The preview proxies to the person's workspace; opening a project starts
  it; the preview says it is being prepared, and implementers can unfold the
  steps and the log.
- `CHATCMS_WORKSPACES_AT_ONCE` and `CHATCMS_WORKSPACE_IDLE`, old names still
  read; NixOS options `workspaceIdle` and `workspacesAtOnce`.
- Moved here from chatcms-9iuz: sharing after each save, conflicts on the
  same field, publish sharing the publisher's saves first.
- Fixed on the way: a person without a display name could not save.
- Acceptance: `test/e2e/33-a-preview-of-ones-own.spec.mjs`, two people, two
  different changes, each sees their own, at both widths.
- Disk use per copy was not measured; objects are hardlinked.
