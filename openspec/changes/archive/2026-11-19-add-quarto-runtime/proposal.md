# Proposal

Bean: [chatcms-q9qu](../../../.beans/chatcms-q9qu--quarto-runtime-outputs-and-artifacts.md)
(Quarto runtime, outputs and artifacts), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

The runtime definition can name any package, but nothing has used it for a
site that is not Hugo. Quarto is the second runtime the briefing asks for,
and it makes things that are not a website: a book is a website and a PDF
from one render. A PDF has nowhere to go today; a publish only knows a
remote.

## What Changes

- **A package may name its nixpkgs revision.** `quarto: { version, nixpkgs }`
  builds that version from that revision, without asking Nixhub. Needed
  because Nixhub's revision for Quarto 1.10.18 pairs it with a pandoc it
  cannot use, while a later nixpkgs revision, with the same version number,
  works.
- **Packages the installation provides** (`CHATCMS_PACKAGES`, the NixOS
  module's `packages`): a name and a store path, used when its version is
  the one asked for. An installation without Nix, and the acceptance tests,
  get Quarto this way.
- **A preview that is not told its address is served at its root.** A
  preview command without `{url}` gets requests with the preview's prefix
  taken off; Hugo, which is told, keeps getting the whole path.
- **A preview may write.** In a micro VM the site is a read-only share; the
  guest now lays a memory-backed overlay over it, so Quarto can write its
  output beside the site, and nothing it writes leaves the VM.
- **A site with a runtime definition can be brought in** without a Hugo
  configuration.
- **Outputs and destinations in the workflow file.** The publish workflow
  may list the build's outputs, each to `publish`, `artifact` or `both`.
  What goes to publish is published as today; artifacts are kept.
- **Artifacts:** kept per project and per publish on disk, recorded in the
  database, downloadable by anybody who may read the project; an Artifacts
  item in the Site navigation after History once a project has any, and the
  publish's own downloads on its Done state.
- **Code cells.** A runtime definition may list Python and R packages; they
  are built with Nix at the pinned revision and handed to Quarto, so Python
  and R cells run in the preview and in the build sandbox, without network
  or secrets.
- **A Quarto book fixture** in `test/fixtures/quarto-book/`: chapters as a
  Quiqr collection, a Python cell, Typst for the PDF, freeze recommended.

## Capabilities

### New Capabilities

- `runtime/code-cells`: Python and R packages for a site's code cells.
- `workflows/outputs`: a build's outputs and where each goes.
- `artifacts/store`: keeping, listing and downloading artifacts.
- `onboarding/runtime`: bringing in a site that declares its runtime.

### Modified Capabilities

- `runtime/definition`: a package's revision, packages the installation
  provides, previews served at their root, and previews that write.

## Impact

- `internal/runtime`, `internal/nixpin`: revisions, provided packages,
  Python and R environments.
- `internal/preview`, `internal/workspace`: rooted previews.
- `nix/workspace-guest.nix`: the overlay over `/site`. NixOS module:
  `packages`.
- `internal/workflows`: `outputs`; the engine hands artifacts to a keeper.
- New `internal/artifacts`, a migration, queries, `/artifacts` endpoints.
- `internal/onboarding`: a runtime definition counts as a site.
- Frontend: the Artifacts page and navigation item, downloads on Done.
- Acceptance (milestone 07, epic 9): the Quarto book previews, publishes its
  website and keeps its PDF from one publish; a PDF preview shows its text.

## Assumptions

- **Quarto from nixos-unstable** (1.10.18, Typst 0.15.1, pandoc 3.7), pinned
  by revision in the fixture. It renders a book's website and its Typst PDF
  offline in about four seconds; the Typst book template is bundled.
- **No live reload for Quarto in the pane.** Quarto's reload script only
  recognises its own proxy prefixes, so under `/v/<key>/` it cannot connect;
  the pane already reloads the preview after every save.
- **The local backend writes into the person's copy.** It has no overlay;
  a Quarto site's own `.gitignore` keeps `.quarto/` and the output out of
  saves, as Quarto's project templates do. The micro VM, which is what
  hosting uses, never writes into the copy.
- **Artifacts show once a project has any**, rather than reading the
  workflow on every navigation; the publish Done state shows that publish's
  own.
- **Artifacts are kept, not pruned**, in this change; a size limit per file
  (200 MB) refuses a runaway build.
- **Python and R name their revision.** `python: { nixpkgs, packages }` and
  `r: { nixpkgs, packages }`; nothing is guessed from the other packages.
  Jupyter is added to Python by ChatCMS, since Quarto needs it to run a
  cell.
- **A preview with pages of its own opens its start page.** The pane follows
  the page being edited only for Hugo, whose addresses it knows; any other
  preview opens the runtime's `open` page.
- **Fixed on the way:** publishing built pages set the author from the
  publish settings even when they named nobody, and git refused the empty
  name; the system identity is used then. A publisher's failure is now in the
  step's technical detail.
