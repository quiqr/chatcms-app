# Tasks

## 1. Shared instance settings

- [x] 1.1 Add `scripts/dev-instance.sh` that validates `CHATCMS_INSTANCE` and sets `instance`, `state`, `database`, `port`, `web_port`, `smtp_port`, `inbox_port` and `host`. Verify with `scripts/dev-scripts.test.sh`: no instance gives 8080/5173, `test` gives 8081/5174, `CHATCMS_PORT` and `CHATCMS_WEB_PORT` override, and `Bad!` is refused.

## 2. Backend script

- [x] 2.1 Rename `scripts/dev.sh` to `scripts/dev-be.sh` and have it source `dev-instance.sh` in place of its own block. Verify that `bash scripts/dev-be.sh` and `CHATCMS_INSTANCE=test bash scripts/dev-be.sh` start on 8080 and 8081 as before.
- [x] 2.2 Change its printed frontend hint and its help header to name `scripts/dev-fe.sh` with the same `CHATCMS_INSTANCE`/`CHATCMS_HOST` prefix. Verify by reading the notice after a start.

## 3. Frontend script

- [x] 3.1 Add `scripts/dev-fe.sh`: it refuses outside the dev shell, runs `npm ci` when `web/node_modules` is missing, exports `CHATCMS_PORT` and `CHATCMS_HOST`, and `exec`s `npx vite --port "$web_port" --strictPort` with any extra arguments. Verify with `scripts/dev-scripts.test.sh`: the stub `npx` receives the right port and environment, and a missing `node_modules` triggers `npm ci`.
- [x] 3.2 Add the backend check: a one-second `curl` against `http://127.0.0.1:$port/`, printing a warning with the matching `dev-be.sh` command when nothing answers. Verify with `scripts/dev-scripts.test.sh`: the warning appears with no listener, and stays away when a stub `curl` answers; a smoke run against a real backend on a spare port confirms it.
- [x] 3.3 Write the help header listing every option and variable. Verify that each variable the script reads appears in it.

## 4. Gate and docs

- [x] 4.1 Add a `dev-scripts` check to `flake.nix` that runs `shellcheck` on the three scripts and runs `scripts/dev-scripts.test.sh`. Verify with `nix flake check`.
- [x] 4.2 Replace `scripts/dev.sh` with `scripts/dev-be.sh` in `AGENTS.md`, `CLAUDE.md` and the comment in `web/vite.config.ts`, and add the `dev-fe.sh` lines to the command list. Verify that `grep -rn 'scripts/dev\.sh' --exclude-dir=archive --exclude-dir=node_modules --exclude-dir=.beans .` finds nothing.
- [x] 4.3 Add the rename (as a breaking change for developers) and the new script to `CHANGELOG.md` under `[Unreleased]`. Verify by reading the entry.

## Workflow follow-up

- Ship with `scripts/ship-change.sh split-dev-scripts "<subject>"`.
- Add `openspec-link` to bean chatcms-byq6 and close it with a Summary of Changes.
