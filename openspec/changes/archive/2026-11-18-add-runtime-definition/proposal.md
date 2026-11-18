# Proposal

Bean: [chatcms-64h1](../../../.beans/chatcms-64h1--runtime-definition-and-one-preview-pane.md)
(runtime definition and one preview pane), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

Previews and builds assume Hugo: its flags are written into both backends and
the build. Quarto, and any site that needs another tool, cannot be served.
And a preview is served from the app's own origin, so a site's JavaScript runs
where the app's session lives, which select-to-edit is about to lean on.

## What Changes

- **`quiqr/runtime.yaml`**, a declarative runtime definition in the site:
  packages and their versions (nixpkgs only), the preview command and type
  (web, pdf, other), the page to open, and the build command and its output.
  Commands are argument lists with placeholders the workspace fills in.
  Packages are found and built the way a project's Hugo version is.
- **Without the file, nothing changes:** Hugo at the project's recorded
  version, with today's commands.
- **Both backends run the declared preview command** instead of their own
  Hugo flags; the micro VM gets it, already filled in, on its command line.
- **Builds run the declared build command** through the runner and publish
  its declared output.
- **Previews on their own origin** (`CHATCMS_PREVIEW_URL`, optional): the
  preview origin serves previews and nothing else; the app origin no longer
  serves them; a preview is reached by an address holding a random key made
  when its workspace started, so no cookie crosses origins.
- **One preview pane** picks its viewer by the preview type: a framed page
  with live reload, the browser's own PDF viewer, or a download.

## Capabilities

### New Capabilities

- `runtime/definition`: what a site may declare about how it is previewed and
  built, and what happens when the declaration is wrong.
- `preview/origin`: where previews are served from and how a frame is let in.
- `preview/pane`: how the preview column shows each kind of preview.

### Modified Capabilities

- `workspace/hugo-version`: a version in the runtime definition comes before
  the one the project records.

## Impact

- New `internal/runtime` (parse, validate, fill in, find packages); the
  workspace `Spec` carries a resolved preview; `sitebuild` takes a resolved
  build.
- `internal/preview`: view keys, the preview host, routing by host.
- `nix/workspace-guest.nix`: the preview unit runs the command it is given;
  `jq` in the guest.
- NixOS module: `previewUrl`.
- Frontend: the preview pane's viewers; the frame's address comes from the
  server.
- Acceptance (milestone 07, epic 7): a Hugo site with a runtime definition
  previews as before, from the separate origin; a site whose preview type is
  pdf shows a PDF in the pane.

## Assumptions

- **The format** (flagged in the briefing as open): YAML, because the rest of
  `quiqr/` is YAML; `packages` as a map of nixpkgs attribute to version; each
  command as a list of words. Shown in design.md.
- **Placeholders, not a shell**: `{host}`, `{port}`, `{url}`, `{reload}`,
  `{source}`, `{output}`, `{cache}`. A word that is not a placeholder is
  passed as it is.
- **The site stays read-only in the VM.** A preview command that writes into
  the site (Quarto's does) is the Quarto epic's to solve, with a writable
  layer over the share.
- **Without `CHATCMS_PREVIEW_URL` previews stay on the app's origin**, as
  today, so development and existing installations keep working.
- **No cookie on the preview origin.** A signed ticket turned into a cookie
  would be a third-party cookie inside the app's frame whenever the preview
  address is on another site, and browsers drop those. The key in the
  address lasts exactly as long as the workspace and is checked against the
  person's right on every request.
- **A changed definition** takes effect the next time the person opens the
  preview, found by comparing the file's hash with the one their workspace
  started with; nothing watches the file.
