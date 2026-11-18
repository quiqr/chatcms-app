# Tasks

## 1. The engine asks again

- [x] 1.1 `Sync` during a stopped exchange returns the open conflict and no
  error, fetching nothing; `Publish` still refuses; proven by
  `TestSyncAsksAgainWhileAnExchangeIsInProgress`.
- [x] 1.2 `Sites.Share` passes the open question on; proven by
  `TestShareStopsOnTheSameField` asking again on a second share.

## 2. The site says what it asks

- [x] 2.1 `sitesync.Standing` includes the questions while one is open;
  proven by a sitesync unit test.
- [x] 2.2 `standingCopy.Sync` asks the open question of whichever copy has it
  without touching the remote; proven by
  `TestAnsweringInALaterRequestFinishesTheExchange` in `internal/app`.

## 3. Acceptance

- [x] 3.1 `test/e2e/34-answering-two-people.spec.mjs`: two people change the
  same title, the second opens the sync screen, sees the question, keeps
  theirs, and the shared version holds it. Runs at 1440 px and 390 px.
- [x] 3.2 `scripts/gate.sh ship` green.

## 4. Changelog

- [x] 4.1 `CHANGELOG.md` under [Unreleased], Fixed: a question about what two
  people changed can be answered later.
