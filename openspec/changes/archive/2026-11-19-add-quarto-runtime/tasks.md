# Tasks

## 1. Packages

- [x] 1.1 A package as `{version, nixpkgs}`; `nixpin.PathAt` with the
  version check; proven by runtime and nixpin tests.
- [x] 1.2 `CHATCMS_PACKAGES` and the NixOS module's `packages`; the resolver
  takes a provided package of the right version first; proven by runtime
  and config tests.
- [x] 1.3 `python:` and `r:`; their environments built and handed over as
  `QUARTO_PYTHON` and `QUARTO_R` to both backends and the build; proven by
  runtime, nixpin, backend and sitebuild tests.

## 2. Previews

- [x] 2.1 Rooted previews: `Endpoint.Rooted`, the prefix taken off; proven by
  manager and preview tests.
- [x] 2.2 The overlay over `/site` in the guest; proven by the hand boot test
  and the release gate.
- [x] 2.3 Onboarding recognises a runtime definition; proven by onboarding
  tests.

## 3. Outputs

- [x] 3.1 `outputs` in the workflow file, with its refusals; proven by parse
  tests.
- [x] 3.2 The engine splits publish and artifact files and hands artifacts
  to a keeper after the last step; proven by engine tests.

## 4. Artifacts

- [x] 4.1 Migration, queries, `internal/artifacts` (keep, list, open, the
  size limit); proven by integration tests.
- [x] 4.2 The publish record carries its artifacts; `/artifacts` endpoints;
  `hasArtifacts` on the project; proven by httpapi and integration tests.
- [x] 4.3 The Artifacts page and navigation item, downloads on Done, strings;
  proven by Vitest.

## 5. The fixture and acceptance

- [x] 5.1 `test/fixtures/quarto-book/`: a book with a Python cell, Typst
  PDF, a model, a runtime definition and a publish workflow with outputs.
- [x] 5.2 e2e: bring the book in, preview a chapter, publish, the website is
  published and the PDF is on Done and under Artifacts and downloads; a
  pdf preview shows its text; at both widths. The flake provides Quarto
  to the e2e.

## 6. Ship

- [x] 6.1 `docs/RUNTIME.md`, `docs/ARCHITECTURE.md`, `docs/DEPLOYING.md`,
  `CHANGELOG.md`; freeze recommended in `docs/RUNTIME.md`.
- [x] 6.2 `scripts/gate.sh release` and `scripts/gate.sh ship` green.
