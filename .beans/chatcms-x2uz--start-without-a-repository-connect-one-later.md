---
# chatcms-x2uz
title: Start without a repository, connect one later
status: completed
type: task
priority: normal
created_at: 2026-11-16T18:40:04Z
updated_at: 2026-11-16T23:10:00Z
openspec-link: openspec/changes/archive/2026-11-16-hosted-until-connected
---

A project started from a template or a Quiqr Desktop export can leave the git address empty. ChatCMS keeps the site in a repository of its own until the implementer connects a real one in project settings, which sends everything so far there.

## Summary of Changes

- A template or Quiqr Desktop export can start with no git address. ChatCMS
  keeps the site in `<state>/hosted/<id>.git` and no key is needed; without an
  address, Continue copies the template in at once.
- Project settings show "Kept in ChatCMS, not connected yet" and connect the
  site to an empty repository the key reaches, sending every version there
  and removing the kept copy (`PUT …/projects/{project}/remote`).
- Publishing a kept site is refused with `refused.site-not-connected` and no
  publish record. Removing a kept project removes its kept repository.
- A project with nothing brought in answers `refused.nothing-brought-in-yet`
  on every screen instead of a server error.
- `docs/DEPLOYING.md` lists `stateDir/hosted/` under what to back up.
