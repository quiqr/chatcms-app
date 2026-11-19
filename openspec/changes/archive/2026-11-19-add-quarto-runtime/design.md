# Design

## Context

See proposal.md. `internal/runtime` resolves packages through
`runtime.Pins` (`nixpin`, Nixhub then `nix build`); both backends run the
resolved preview; `sitebuild` runs a declared build through the runner; the
publish workflow builds and hands `sitebuild.Pages` to the publisher.

## Decisions

### 1. The definition

```yaml
packages:
  quarto: { version: "1.10.18", nixpkgs: 02c087f3fe673c9e317f206a627cb6b910336dbe }
python:
  nixpkgs: 02c087f3fe673c9e317f206a627cb6b910336dbe
  packages: [pandas]
preview:
  type: web
  command: [quarto, preview, "{source}", --host, "{host}", --port, "{port}", --no-browser, --to, html]
build:
  output: _book
  command: [quarto, render, "{source}"]
```

A package is a string (a version) or a map with `version` and `nixpkgs`.
`Resolved` gains `Env` (`QUARTO_PYTHON`, `QUARTO_R`) and `Rooted` (the
preview command has no `{url}`).

### 2. Where a package comes from, in order

1. `CHATCMS_PACKAGES` (`quarto=/nix/store/…-quarto-1.10.18,…`): used when the
   store path's name ends in `-<version>`.
2. A pinned revision: `nixpin.PathAt(name, version, rev)` builds
   `github:NixOS/nixpkgs/<rev>#<name>`, checks the store path's version, and
   remembers it under the version `<version>@<rev>`.
3. Nixhub, as before. Hugo still falls back to the installation's Hugo.

Python and R: `nix build --impure --expr` over
`(builtins.getFlake "github:NixOS/nixpkgs/<rev>").legacyPackages.<system>`,
`python3.withPackages (ps: [ps.jupyter ps.<each>])` and
`rWrapper.override { packages = with rPackages; [ <each> ]; }`, names
checked against `^[A-Za-z0-9_-]+$` before they reach the expression.

### 3. Rooted previews

`workspace.Endpoint` gains `Rooted`; the manager sets it from the resolved
preview. `preview.decorate` takes the prefix off the request's path (and
`RawPath`) before proxying when it is set. Live reload for such previews
is not attempted.

### 4. A writable site in the guest

The `site` share mounts at `/site-lower`, read-only; `/site` is an overlay
with `lowerdir=/site-lower`, and upper and work directories on a tmpfs of
512 MB. The overlay is an fstab entry ordered after both, so the preview
unit's `local-fs.target` waits for it. virtiofs as an overlay lower needs no
extra option on this kernel; the hand boot test proves it.

### 5. Outputs

```yaml
on: publish
steps:
  - action: build
  - action: publish
outputs:
  - path: .
    to: publish
  - path: "*.pdf"
    to: artifact
```

`Workflow.Outputs []Output{Path, To}`. A file belongs to an output when the
output is `.`, when the path equals the file or is a directory above it, or
when `path.Match` matches the whole relative path. The engine builds two
views of the build's pages: the publish view (files of publish outputs)
and the artifact list; after the last step it calls
`Keeper.Keep(ctx, run, files)` and puts what was kept on the outcome.

### 6. Artifacts

Table `artifacts (id, project_id, publish_id, name, size, sha256,
created_at)`; files at `<state>/artifacts/<project>/<id>`. The publish
service passes the publish id in `workflows.Run` and returns the kept
artifacts with the record. Endpoints, under the project:

- `GET  …/artifacts` → newest first, 100 at most;
- `GET  …/artifacts/{id}` → the file, `Content-Disposition: attachment`.

The project view gains `hasArtifacts`. The Site navigation shows
Artifacts after History when it is true; the publish Done state lists the
record's artifacts as download links.

### 7. Bringing a site in

`onboarding.Inspect` reads `quiqr/runtime.yaml` first. When it is there and
readable, the site is a site of its runtime (`runtime` in the report,
`quarto` for a definition whose preview program is `quarto`, else the
program's name), and the model is read as for Hugo.
