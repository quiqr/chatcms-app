---
# chatcms-j31v
title: test suite tiers and a parallel e2e suite
status: in-progress
type: task
created_at: 2026-11-17T21:43:44Z
updated_at: 2026-11-17T21:43:44Z
---

The gate runs the whole e2e suite (165 tests, one at a time, about 12 minutes) and the NixOS VM test on every shipped change. That is too long to wait for at every change.

- Run the e2e tests in parallel; each already has a ChatCMS of its own.
- Tiers: a fast tier while working, a ship tier for every change, and the full gate for a release.
