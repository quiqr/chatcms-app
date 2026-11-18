---
# chatcms-9iuz
title: Changes from everyone come together
status: todo
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-18T12:51:48Z
parent: chatcms-z0p7
blocked_by:
    - chatcms-gkdd
---

## Purpose

With a working copy per person, changes have to come together without anyone
seeing a merge conflict. Publishing stays idiot proof.

## Scope

- Small auto-commits per user branch, rebased onto the shared version in the
  background.
- Publish takes everyone's saved changes.
- A real conflict is a plain choice: keep mine or keep theirs. "Let the AI
  combine them" is added by the chat epic.
- Optional page-level soft locking, a project setting that is off by default.
- Reuses the conflict surfacing from chatcms-8d4z where it fits.

## Done when

- Two people edit different pages, and one publish carries both.
- Two people edit the same field: the second is offered a plain choice, with
  no git words, at both widths.

## Open decisions

- Final merge strategy. Default: rebase each person's branch onto main after
  every save. Flag in the proposal.

## References

`docs/BRIEFING-M07.md`, "Collaboration / merging". Copy rules in
`docs/design/DESIGN.md` section 7.
