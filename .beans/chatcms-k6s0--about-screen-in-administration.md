---
# chatcms-k6s0
title: about screen in administration
status: completed
type: feature
created_at: 2026-11-17T22:10:02Z
updated_at: 2026-11-18T09:00:00Z
openspec-link: openspec/changes/archive/2026-11-17-about-screen
---

An About screen in the platform administration: the version that is running, what it runs on, the license, the GitHub page and the credits (the Quiqr Team and its contributors; anonymous for now, GitHub names later).

## Summary of Changes

Shipped with `about-screen` (commit 42fa83e):

- Administration, About, after History, for platform admins only, backed by
  `GET /api/administration/about`.
- Version and the source it was built from ("0.1.0 · a3bab43", ", changed",
  or "development build"); `chatcmsd --version` says the same. Only the
  deployed package knows its source, so the checks stay cached per commit.
- What it runs on: Hugo, git (through gitsync), PostgreSQL and Go, each with
  a two-second limit and "unknown" when one does not answer.
- The Apache 2.0 license with its full text, links to GitHub, the changelog
  and the issues, and credits to the Quiqr Team with a link to the
  contributors on GitHub.
