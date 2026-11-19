---
# chatcms-q9qu
title: Quarto runtime, outputs and artifacts
status: completed
type: epic
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-19T07:19:37Z
openspec-link: openspec/changes/archive/2026-11-19-add-quarto-runtime
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

## Summary of Changes

Shipped as `openspec/changes/archive/2026-11-19-add-quarto-runtime`.

- Quarto 1.10.18 from nixos-unstable, pinned by revision: a runtime package
  may be `{version, nixpkgs}`, because Nixhub's revision for that version
  pairs Quarto with a pandoc it cannot use. An installation may provide
  packages itself (`CHATCMS_PACKAGES`, `services.chatcms.packages`).
- A preview command without `{url}` is served at its root; a non-Hugo
  preview opens its own start page. In a micro VM `/site` is a tmpfs overlay
  and the preview runs as the site's owner, so Quarto writes and the
  person's copy stays untouched (real VM test: answers after about 10 s).
- Python and R packages for code cells are built with Nix at a named
  revision and handed over as `QUARTO_PYTHON` and `QUARTO_R`; cells run in
  the workspace and in the build sandbox, offline.
- Workflow `outputs` send parts of a build to publish, artifact or both.
  Artifacts are kept per publish (`internal/artifacts`, table `artifacts`),
  listed under Artifacts after History once a project has any, and offered
  on the publish Done state.
- A site with `quiqr/runtime.yaml` can be brought in without a Hugo
  configuration.
- Fixture `test/fixtures/quarto-book/` and e2e `38-a-quarto-book`: bring it
  in, preview, publish the website with its Python output, keep and download
  the Typst PDF, preview as a PDF through PDF.js; both widths.
- Fixed on the way: publishing built pages failed when the settings named no
  author; a publisher's error now reaches the technical detail.
- A Quarto template for the gallery is left for later, as agreed.
