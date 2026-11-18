# Proposal

Bean: [chatcms-a9sj](../../../.beans/chatcms-a9sj--answering-what-two-people-changed-finds-no-questio.md)
(answering what two people changed finds no question once the exchange has
stopped), milestone 07.

## Why

When an exchange stops because two people changed the same field, nothing
can pick the question up again. `Sync` answers a stopped exchange with
`ErrExchangeInProgress` and no conflict, so answering finds nothing to
answer, "Check for updates" fails, and how the site stands says "asking"
without saying what. Since every person has their own copy (chatcms-gkdd),
this stops a person's saves from ever reaching the shared version.

## What Changes

- A sync during a stopped exchange reports the open question again, without
  fetching or replaying anything. Publishing still refuses.
- `Sites.Share` passes that question on instead of an error.
- How a site stands includes the questions while one is open.
- Answering applies the answers and finishes the exchange, in a person's copy
  and in the shared copy; the person's work then reaches the shared version.

## Capabilities

### New Capabilities

### Modified Capabilities

- `sync/settling`: an open question can be read again and answered at any
  later moment, not only in the request that stopped the exchange.

## Impact

- `internal/gitsync` (`Sync`, `Share`), `internal/sitesync` (`Standing`),
  `internal/app/copies.go` (`standingCopy.Sync`).
- Seam: `gitsync`. Nothing about branches or paths leaves it; the change is in
  what `Sync` reports.
- Acceptance: a new e2e test where two people change the same field and the
  second answers "Keep mine".

## Assumptions

- Asking how a site stands while a question is open may run git to read it
  again; that only happens while the site is asking, not on every poll.
