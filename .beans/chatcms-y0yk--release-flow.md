---
# chatcms-y0yk
title: release flow
status: completed
type: task
priority: normal
created_at: 2026-11-16T11:12:53Z
updated_at: 2026-11-16T18:00:00Z
openspec-link: openspec/changes/archive/2026-11-16-release-flow
parent: chatcms-27ng
---

- release scripts
- release docs
- first release

Look at /home/pim/gh.nivis-project/nivis-openspec-stores/nivis/openspec/changes/archive/2026-06-15-release-management and adapt this change to this project

## Summary of Changes

- `VERSION` is the only place the version is written; Nix and `go build` read it.
- `scripts/release.sh patch|minor|major [--dry-run]` refuses unsafe starts,
  bumps, dates the changelog, gates, tags, pushes and creates the GitHub
  release. Tested with stubs in the `release` flake check.
- `docs/RELEASING.md` rewritten around the script.
- Changelog entries are held to two lines of 80 characters by a gate check.
- First release: v0.1.0, https://github.com/quiqr/chatcms-app/releases/tag/v0.1.0
