---
# chatcms-byq6
title: docs for running locally for dev
status: completed
type: task
priority: normal
created_at: 2026-11-16T11:10:08Z
updated_at: 2026-11-16T15:30:00Z
openspec-link: openspec/changes/archive/2026-11-16-split-dev-scripts
parent: chatcms-27ng
---

- how to start db/frontend/backend
  - maybe a one stop script
- development docs for nix users

## Summary of Changes

- `scripts/dev.sh` is now `scripts/dev-be.sh`; it starts database, Mailpit,
  seed and server as before.
- New `scripts/dev-fe.sh` starts Vite, installs dependencies on a fresh
  checkout, follows `CHATCMS_INSTANCE` and `CHATCMS_HOST`, and warns when no
  backend answers.
- `scripts/dev-instance.sh` holds the instance-to-port mapping both share.
- `dev-scripts` flake check: shellcheck plus stub-based tests.
- README, AGENTS.md and CLAUDE.md name the two scripts.
- Not done: a single script that starts both halves.
