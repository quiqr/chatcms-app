# Proposal

Bean: [chatcms-byq6](../../../.beans/chatcms-byq6--docs-for-running-locally-for-dev.md)
(docs for running locally for dev), milestone chatcms-27ng (05 develop
happiness).

## Why

The README promises two scripts, `scripts/dev-be.sh` for the backend and
`scripts/dev-fe.sh` for the frontend. Only one script exists, and it starts
the backend. The frontend is started by hand with `cd web && npm run dev`.
For a second instance the user also has to type the API port and the Vite
port themselves, and nothing tells them when the backend is not running.

## What Changes

- **BREAKING** `scripts/dev.sh` is renamed to `scripts/dev-be.sh`. Its
  behaviour stays the same. Only its name and the frontend hint it prints
  change.
- A new `scripts/dev-fe.sh` starts the Vite development server. It:
  - refuses outside `nix develop`
  - installs the frontend dependencies when `web/node_modules` is missing
  - works out the API port and its own port from `CHATCMS_INSTANCE`
  - passes `CHATCMS_HOST` on
  - warns and starts anyway when no backend answers on the API port
- One small file that both scripts source holds the mapping from instance name
  to ports, so the two halves cannot disagree about it.
- `AGENTS.md`, `CLAUDE.md`, `web/vite.config.ts` and the help text in both
  scripts name the new scripts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `platform/local-development`: adds requirements for starting the frontend
  with one script that follows the backend's instance and host, and that warns
  when the backend is missing.

## Impact

- `scripts/dev.sh` (renamed), `scripts/dev-fe.sh` (new),
  `scripts/dev-instance.sh` (new).
- `AGENTS.md`, `CLAUDE.md`, `web/vite.config.ts`, `CHANGELOG.md`.
- `flake.nix` gets a check that runs the scripts' tests.
- No Go, database or frontend code changes. No seam (gitsync, runner, policy,
  auth) is touched.

## Acceptance test served

None from `docs/BRIEFING-M01.md` section 6 directly. This is developer
tooling. It makes the acceptance suite easier to run by hand against a second
instance.

## Assumptions

- The instance ports stay as they are: the usual instance uses API 8080 and
  Vite 5173, and a named instance uses API 8081 and Vite 5174. `CHATCMS_PORT`
  and a new `CHATCMS_WEB_PORT` override them.
- The backend check only asks whether anything answers on the API port. The
  server has no health endpoint, and this change does not add one.
- Arguments after `dev-fe.sh` are passed on to Vite unchanged.
- The bean's mention of a "one stop script" for database, backend and frontend
  together is out of scope here. The two scripts are the building blocks for
  it.
