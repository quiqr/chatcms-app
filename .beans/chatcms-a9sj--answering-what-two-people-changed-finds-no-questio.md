---
# chatcms-a9sj
title: Answering what two people changed finds no question once the exchange has stopped
status: completed
type: bug
priority: high
created_at: 2026-11-18T14:50:02Z
updated_at: 2026-11-18T15:11:53Z
openspec-link: openspec/changes/archive/2026-11-18-answer-a-stopped-exchange
parent: chatcms-z0p7
---

## What happens

`sitesync.Service.Answered` calls `site.Sync` to get the question it is
answering. When an exchange has stopped on a conflict, `gitsync` `Sync`
returns `ErrExchangeInProgress` with an empty result, so `Answered` sees no
conflict and returns the standing without applying the answers. The stopped
exchange stays stopped.

## Why it matters now

Since chatcms-gkdd every person has their own copy, and two people changing
the same field stop that person's share until it is answered. Answering is
the only way forward short of giving up.

## Expected

`Answered` gets the open question, applies the answers and finishes the
exchange, in a person's copy and in the shared copy.

## Notes

No e2e test answers a question; add one (two people, same field, answer
keep mine, both previews and the shared version agree). Found while shipping
`openspec/changes/archive/2026-11-18-add-workspace-per-person`.

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-18-answer-a-stopped-exchange`.

- A sync during a stopped exchange reports the open question again, fetching
  nothing; publishing still refuses.
- `Sites.Share` passes the question on instead of an error.
- How a site stands includes the questions while one is open, so the sync
  screen shows them after a reload and "Check for updates" no longer fails.
- A person's stopped exchange is asked of their copy without touching the
  remote; answering finishes it and shares their work.
- Acceptance: `test/e2e/34-answering-two-people.spec.mjs`, a title changed
  here and elsewhere, the question still asked after coming back, "Keep
  mine" kept, at both widths.
