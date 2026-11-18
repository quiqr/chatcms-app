---
# chatcms-q9qu
title: Quarto runtime, outputs and artifacts
status: todo
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-18T21:47:03Z
parent: chatcms-z0p7
blocked_by:
    - chatcms-5odk
---

## Purpose

Quarto as the second runtime, with PDF output, and outputs that go somewhere:
stored as artifacts, published, or both. Supersedes chatcms-wmm1.

## Scope

- Quarto with Typst as PDF engine (TeX Live optional), all project types:
  website, book, manuscript, presentation.
- `quarto preview` in the workspace; the final PDF as a one-shot build through
  `runner.Runner`.
- Outputs and destinations declared in the workflow file: artifact, publish
  or both.
- Artifacts stored in ChatCMS, versioned and downloadable; the artifacts panel
  shows only when a project defines them (Site navigation after History, and
  on the publish Done state, per `docs/design/DESIGN.md` section 8).
- Python and R code cells run inside the sandbox with no network and no
  secrets; their dependencies pinned through Nix in the runtime definition.
- Quarto's freeze recommended in docs and templates, not enforced.

## Done when

- A Quarto project renders a PDF preview and produces a PDF artifact.
- A Quarto book publishes its website and stores its PDF from one publish.

## References

`docs/BRIEFING-M07.md` step 4 and the addendum "Quarto support".
