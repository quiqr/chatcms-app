---
# chatcms-koz4
title: The e2e harness can lose its port to another process before the server binds
status: todo
type: bug
priority: normal
created_at: 2026-11-18T14:50:02Z
updated_at: 2026-11-18T14:50:02Z
---

## What happens

The harness asks the system for a free port, releases it and starts the
server on it later. In between, another process can take the port: a Hugo
started by a workspace in a parallel test does the same dance. The server
then stops with `bind: address already in use` and the test fails before it
starts. Seen once in the release gate on
`24-platform-admin-area.spec.mjs` ("one organisation is opened, limited and
deleted").

## Why it got likelier

Since chatcms-gkdd opening a project starts a workspace, so far more Hugo
servers start during a run than before.

## Expected

A test never fails on a port it did not choose. Options: let the server bind
port 0 and report the address it got, or retry a start that failed on the
bind.
