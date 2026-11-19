---
# chatcms-9iuz
title: Changes from everyone come together
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:48Z
updated_at: 2026-11-19T10:07:29Z
openspec-link: openspec/changes/archive/2026-11-19-add-ai-combine
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

## Summary of Changes

- Saves come together in the background (chatcms-yr88, delivered by the
  per-person workspace change); the same field changed twice is a plain
  question with no git words, at both widths (e2e 34).
- The AI can combine two versions of a field (chatcms-izhr, `openspec/changes/archive/2026-11-19-add-ai-combine`).
- Two people editing different pages are published together: proven by e2e
  `40-two-people-one-publish`, added after this epic's changes shipped.
- Soft page locking, optional here, is left out and filed as chatcms-na8c.
