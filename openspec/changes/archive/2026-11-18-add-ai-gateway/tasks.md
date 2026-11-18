# Tasks

## 1. Storage

- [x] 1.1 Migration `00020_ai.sql` (`ai_tokens`, `ai_usage`, organisation
  columns) and queries; `sqlc diff` clean.

## 2. The gateway

- [x] 2.1 `internal/aigateway` tokens: issue, revoke, look up, hashed; proven
  by unit tests on a memory store and an integration test on Postgres.
- [x] 2.2 The handler: address, token, route, model rewrite, key in place of
  the token, non-streamed and streamed answers, OpenAI-shaped refusals;
  proven by tests against an `httptest` provider.
- [x] 2.3 Limits: requests a minute in memory, tokens a month from usage;
  refused with plain reasons; proven by unit tests with a fake clock and an
  integration test of the monthly sum.
- [x] 2.4 Usage recorded for every answered request, streamed included;
  proven by the handler tests and the Postgres integration test.

## 3. Workspaces carry a token

- [x] 3.1 `workspace.Settings.Tokens`: issued before start, revoked on sleep,
  failed start and reconcile; `Spec.Token`; proven by manager tests.
- [x] 3.2 The microvm backend puts token and gateway on the command line; the
  guest writes `/run/chatcms/ai.env`; proven by the backend test and the
  hand-run boot test.

## 4. Admin

- [x] 4.1 `GET`/`PUT /api/administration/organisations/{org}/ai`, platform
  admin only; proven by httpapi and integration tests.
- [x] 4.2 "AI use" on the organisation page: limits and this month's use;
  strings in `en.ts` and `nl.ts`; proven by Vitest.

## 5. Configuration and deployment

- [x] 5.1 `CHATCMS_AI_*` variables and the gateway listener in the server;
  proven in `config_test.go` and an integration test that calls the
  gateway on its listener.
- [x] 5.2 NixOS module: `ai` options, key credential, port 8091 on `chv*` in
  nftables and the NixOS firewall; the VM test checks both;
  `scripts/gate.sh release` green.
- [x] 5.3 `docs/ARCHITECTURE.md`, `docs/DEPLOYING.md`, `CHANGELOG.md`.

## 6. Ship

- [x] 6.1 `scripts/gate.sh ship` green.
