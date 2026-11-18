---
# chatcms-a9sj
title: Answering what two people changed finds no question once the exchange has stopped
status: in-progress
type: bug
priority: high
created_at: 2026-11-18T14:50:02Z
updated_at: 2026-11-18T14:54:33Z
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
