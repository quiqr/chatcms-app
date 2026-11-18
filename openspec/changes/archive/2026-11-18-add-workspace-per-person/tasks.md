# Tasks

## 1. A person's copy in gitsync

- [x] 1.1 `Sites.OpenFor` makes a person's copy from the shared copy on first
  use (`clone --local`, origin the shared copy, never a key) and opens it
  afterwards; proven by `TestOpenForMakesACopyOnce`,
  `TestOpenForRefusesUnusableNames` and `TestPersonCopyNeedsNoKey`.
- [x] 1.2 The shared copy accepts sends with
  `receive.denyCurrentBranch=updateInstead`, set the first time a person's
  copy is opened; proven by `TestSharedCopyTreeFollowsASend`.
- [x] 1.3 `Sites.Share` brings a person's copy level and sends it, taking the
  shared lock before the person's, retrying a lost race up to three times;
  proven by `TestShareReachesTheSharedVersion`,
  `TestShareKeepsAnotherPagesSave` and `TestTwoSharesRacing` (race detector).
- [x] 1.4 A conflict in `Share` stops with a `Conflict` and leaves the shared
  version untouched; proven by `TestShareStopsOnTheSameField`.
- [x] 1.5 `OpenFor` brings a copy with no exchange in progress level with the
  shared version; proven by `TestRemoteArrivalReachesACopyOnOpen`.
- [x] 1.6 `Sites.Unshared` and `Sites.RemoveFor`; proven by unit tests, and
  `internal/gitsync` stays above 80% coverage.
- [x] 1.7 The source test that forbids git outside gitsync still passes; the
  `sync/repository` "only the repository component runs git" scenario is
  green.

## 2. Features read and write the person's copy

- [x] 2.1 `content.Sources.Site` takes the person; the app adapter calls
  `OpenFor`; `internal/content` tests updated and green.
- [x] 2.2 The same for `media`; tests green.
- [x] 2.3 The same for `history`, going back included; tests green.
- [x] 2.4 The same for `sitesync`: standing and settling read the person's
  copy, and settling a person's stopped exchange settles it in that copy;
  proven by `TestStandingCountsWhatIsNotShared`,
  `TestStandingBringsInWhatOthersSharedWhileOffline`,
  `TestAnsweringSettlesThePersonsCopyAndShares` and
  `TestGivingUpRestoresThePersonsCopy` in `internal/app`.
- [x] 2.5 Saving queues `site-share` for (project, person) instead of
  `site-sync`; the job runs `Share` and then queues `site-sync`; proven by
  `TestSaveReachesTheSharedVersionByJob` (integration).
- [x] 2.6 Publish calls `Share` for the publishing person before it reviews
  or publishes the shared copy, and refuses while a question waits; proven by
  `TestPublishSharesThePublishersSavesFirst`,
  `TestPublishStopsWhenSharingFails` and
  `TestAnsweringSettlesThePersonsCopyAndShares`.
- [x] 2.7 `internal/httpapi` stays above 80% coverage; `go test ./...` and
  `go test -tags=integration ./...` green.

## 3. The workspace package and the local backend

- [x] 3.1 Migration `00018_workspaces.sql` and sqlc queries (reserve a place
  under the cap, mark warm, mark sleeping, touch, list for reconcile and
  cleanup); `sqlc diff` clean; proven by a migration up and down test.
- [x] 3.2 `internal/workspace`: `Key`, `Spec`, `Endpoint`, `Handle`,
  `Backend`, and the `Workspaces` manager with start, steps and the in-memory
  log ring; tested against a fake backend in
  `internal/workspace/workspacetest`.
- [x] 3.3 The cap is taken in one transaction; proven by
  `TestTheLastPlaceIsTakenOnce` (integration, two goroutines).
- [x] 3.4 The sweep: sleep after idle, cleanup hourly (person without read,
  project archived, 30 days unused with nothing unshared); proven by tests
  that move the manager's clock.
- [x] 3.5 Reconcile at start-up: alive becomes warm, anything else is stopped
  and sleeping; proven by `TestReconcileReattaches` and
  `TestReconcileStopsWhatIsGone`.
- [x] 3.6 The `local` backend: Hugo in its own process group, handle
  `pid:port:hugo`, `Alive` checks the pid, its cmdline and the port; proven
  by `TestLocalAliveRefusesAReusedPid` and the existing Hugo-backed preview
  tests moved here.
- [x] 3.7 App stop leaves warm workspaces of a backend that runs apart
  running, and stops and records as sleeping those of the local backend;
  proven by manager tests with the fake backend both ways, and by
  `TestALocalWorkspaceEndsWithTheApp`.

## 4. Preview behind the workspace

- [x] 4.1 `internal/preview` keeps the policy check, the proxy and the
  live-reload rewrite, and asks the manager for an `Endpoint` keyed on
  (person, project); `Servers` is removed; the `preview/serving` and
  `preview/lifetime` scenarios are green in `preview_test.go`.
- [x] 4.2 A refusal at the cap still carries
  `refused.too-many-sites-being-previewed`; proven by the existing cap test.
- [x] 4.3 `internal/workspace` and `internal/preview` covered by the 70%
  overall gate.

## 5. Configuration and deployment

- [x] 5.1 `CHATCMS_WORKSPACES_AT_ONCE` and `CHATCMS_WORKSPACE_IDLE`, falling
  back to the preview names with a warning; proven in `config_test.go`.
- [x] 5.2 The NixOS module gains `workspaceIdle` and `workspacesAtOnce` (the
  module never had preview options, so nothing is renamed); `scripts/gate.sh
  release` run by hand, the `nixos-vm` check green.
- [x] 5.3 `docs/ARCHITECTURE.md` describes `internal/workspace`, the
  person's copy and the share job; `docs/DEPLOYING.md` names the new
  variables.

## 6. Starting and the preparing indicator

- [x] 6.1 `POST` and `GET /api/orgs/{org}/projects/{project}/workspace`; the
  steps and log only with `ActionSeeTechnicalDetail`; proven by httpapi tests
  for an implementer, an editor and a viewer.
- [x] 6.2 Opening a project (any project screen or the editor) starts the
  workspace; proven by the "starting a workspace" Vitest tests.
- [x] 6.3 `Preview.tsx` shows the preparing indicator while starting, the
  preview once warm, and a reason when the start failed; the unfoldable steps
  and log only for implementers and owners; Vitest tests for each state.
- [x] 6.4 Every new string in `web/src/strings/en.ts` and `nl.ts`; `npm run
  check:strings` and `npm run typecheck` green; no forbidden word in what an
  editor sees, proven by the strings test.

## 7. Acceptance

- [x] 7.1 `test/e2e/33-a-preview-of-ones-own.spec.mjs`: two people open the
  same project in two browsers, each saves a different title on the same
  page, and each preview shows their own title and not the other's, also
  while the second is asked about the difference. Runs at 1440 px and 390 px.
- [x] 7.2 `05-live-preview.spec.mjs` still green: a saved change shows in the
  preview without sleeping.
- [x] 7.3 An e2e test that the preparing indicator shows on a cold open and
  that an implementer can unfold the steps.
- [x] 7.4 `scripts/gate.sh ship` green.

## 8. Changelog

- [x] 8.1 `CHANGELOG.md` under [Unreleased]: a preview of one's own per
  person; saves reach everyone by themselves; the preparing indicator; the
  renamed configuration. Each entry at most two lines of 80 characters.

## Workflow follow-up

- Ship with `bash scripts/ship-change.sh add-workspace-per-person "<subject>"`.
- Add `openspec-link` to chatcms-gkdd and close it with a `## Summary of
  Changes`, naming the disk use measured in 1.x and what moved from
  chatcms-9iuz into this change.
