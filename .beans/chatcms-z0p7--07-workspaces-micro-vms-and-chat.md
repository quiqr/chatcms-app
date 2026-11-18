---
# chatcms-z0p7
title: 07 Workspaces, micro VMs and chat
status: todo
type: milestone
priority: high
created_at: 2026-11-18T12:50:44Z
updated_at: 2026-11-18T12:50:44Z
---

## Goal

A user opens a project and gets their own workspace: a working copy on their
own branch, a preview (Hugo or Quarto) and an OpenCode chat agent. An edit
asked for in chat appears in the preview and is saved as a version on that
user's branch.

## Why it exists now

Milestone 01 runs one Hugo server per project on one shared working copy, and
there is no agent runtime. A chat agent runs code the platform does not trust,
so it needs an isolated place to run, and isolation needs a seam to sit
behind. `internal/runner.Runner` is that seam for one-shot builds; this
milestone adds a Workspace seam for long-running work: preview and agent.

## Order

The epics are chained: each is blocked by the one before it and is testable
on its own, so `beans list --ready` always names the next one.

1. Workspace per user per project, local backend
2. Changes from everyone come together
3. Micro VM workspace backend
4. gVisor workspace backend
5. AI gateway on the host
6. Chat with OpenCode in the workspace
7. Runtime definition and one preview pane
8. Quarto runtime, outputs and artifacts
9. Select-to-edit in the preview

## Definition of done

- Two users preview the same project at once, each with their own changes.
- A preview runs inside a micro VM with the project's pinned Hugo version.
- "Change X on page Y" in chat changes the file, shows in the preview and is
  committed on the user's branch.
- A Quarto project renders a PDF preview and produces a PDF artifact.
- No content creator ever sees a merge conflict or a git word.

## Out of scope

vsock networking, remote micro VM hosts, customer self-hosted runners,
advanced runtime definitions (custom flakes), the Firecracker build runner
(chatcms-uwdh) and autoscaling. OpenSpec guardrails are filed as a draft epic
under milestone 06.

## Source

`docs/BRIEFING-M07.md`, copied from Pim's briefing. Its "Open decisions"
section lists the defaults each proposal picks and flags.
