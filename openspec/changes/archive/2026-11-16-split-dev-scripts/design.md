# Design

## Context

`scripts/dev.sh` works out an instance's state directory, database, API port
and mail ports in one block, then prints a hint for starting the frontend by
hand. `web/vite.config.ts` reads `CHATCMS_PORT` for the proxy target and
`CHATCMS_HOST` for its listen address. Vite's own port is set with `--port`.

## Goals / Non-Goals

**Goals:**

- Both scripts get the instance-to-port mapping from one place.
- `dev-fe.sh` stays a thin wrapper: it sets the environment and then hands
  over to Vite with `exec`, so Ctrl-C reaches Vite directly.

**Non-Goals:**

- A single script that starts database, backend and frontend together.
- A health endpoint on the server.

## Decisions

### A sourced `scripts/dev-instance.sh` holds the mapping

It validates `CHATCMS_INSTANCE` and sets `instance`, `state`, `database`,
`port`, `web_port`, `smtp_port`, `inbox_port` and `host`. `dev-be.sh` sources
it in place of its own block. `dev-fe.sh` sources it and uses only `port`,
`web_port` and `host`.

Alternative: copy the few lines into `dev-fe.sh`. That is less code, but the
two scripts can then drift apart, which is the exact problem the frontend hint
works around today.

### The backend check is a one-second `curl` against the API port

Any HTTP answer, a 404 included, counts as running. `curl` is already in the
dev shell and `dev-be.sh` already uses it for Mailpit. A plain TCP probe
(`/dev/tcp`) would also work, but it does not show that the thing on the port
speaks HTTP.

### Dependencies are installed with `npm ci` only when `node_modules` is missing

This is quick on every start after the first. A stale `node_modules` after a
`package-lock.json` change is left to the developer: `npm ci` on every start
would be too slow.

### The scripts are tested with stubs on `PATH`

`scripts/dev-scripts.test.sh` runs `dev-fe.sh` with stub `npm`, `npx` and
`postgres` commands that write down their arguments and environment. Then it
checks the ports, the host, the warning and the refusal of a bad instance
name. The test runs as a flake check together with `shellcheck` on the three
scripts.

### The help text stays at the top of each script

The README tells people to read the scripts for their options, so the header
block is the documentation the spec asks for. The script bodies carry no
comments.

## Risks / Trade-offs

- [Renaming `dev.sh` breaks habits and shell history] → The README already
  names the new scripts. `CHANGELOG.md` records the rename.
- [Something other than ChatCMS answers on the API port and hides the warning]
  → `dev-be.sh stop` already names whatever holds the port. The warning only
  helps and was never a guarantee.
