# The runtime definition

A site may say how it is previewed and built in `quiqr/runtime.yaml`. A site
without one is previewed and built by Hugo at the project's recorded version,
as before.

```yaml
packages:
  quarto: { version: "1.10.18", nixpkgs: 02c087f3fe673c9e317f206a627cb6b910336dbe }
python:
  nixpkgs: 02c087f3fe673c9e317f206a627cb6b910336dbe
  packages: [pandas]
preview:
  type: web
  open: /
  command: [quarto, preview, "{source}", --host, "{host}", --port, "{port}", --no-browser, --to, html]
build:
  output: _book
  command: [quarto, render, "{source}"]
```

`test/fixtures/quarto-book/` is a whole Quarto book set up this way: chapters
as a Quiqr collection, a Python cell, a Typst PDF, and a publish workflow that
publishes the website and keeps the PDF.

## packages

A package is a version, or a version and the nixpkgs revision to build it
from. Their `bin` directories make up the command's `PATH`. A package comes
from, in this order:

1. the installation itself, when it provides that name at that version
   (`CHATCMS_PACKAGES`, the NixOS module's `services.chatcms.packages`);
2. the named nixpkgs revision, built with Nix, checked to hold that version;
3. Nixhub's revision for that version, built with Nix.

Building needs an installation with Nix (`CHATCMS_NIX_BIN`). Without it only
provided packages and the installation's Hugo can be had.

Name the revision when Nixhub's is wrong for you. Quarto 1.10.18 is such a
case: Nixhub's revision pairs it with a pandoc it cannot use, while
nixos-unstable's build of the same version works.

## python and r

Code cells in Quarto run with a Python and an R of their own when the
definition lists packages for them:

```yaml
python: { nixpkgs: <revision>, packages: [pandas, matplotlib] }
r: { nixpkgs: <revision>, packages: [ggplot2] }
```

ChatCMS adds Jupyter to Python, and knitr and rmarkdown to R, builds each as
one environment with Nix and hands it to Quarto as `QUARTO_PYTHON` and
`QUARTO_R`. Without these lists Quarto uses the Python and R it comes with.

Cells run inside the workspace while previewing, and inside the build
sandbox while publishing, with no network and no secrets. A cell that needs
data has to find it in the site.

## Freeze

Quarto's freeze (`execute: freeze: auto` in `_quarto.yml`) keeps the output
of cells that did not change, so a publish does not run them again. ChatCMS
recommends it for any site whose cells are slow, and leaves it to the site:
the `_freeze/` directory is the site's to commit.

## What a preview writes

A preview may write beside the site, as `quarto preview` does. In a micro VM
it writes into memory and nothing reaches the person's copy. The local
backend writes into the copy, so a site's `.gitignore` has to keep such
output out, as Quarto's own templates do with `/.quarto/` and the output
directory.

## preview

| Field     | Meaning                                                        |
| --------- | -------------------------------------------------------------- |
| `type`    | `web` (the default), `pdf` or `other`                          |
| `open`    | The page the pane opens, starting with `/`; `/` by default     |
| `command` | A list of words; the first is a program from a declared package |

A `web` preview follows the page being edited when it is Hugo; any other
preview opens `open`. A command without `{url}` is served at its own root:
ChatCMS takes the preview's prefix off each request. A `pdf` preview shows `open`
in the pane. An `other` preview offers `open` to open in a tab of its own.

## build

| Field     | Meaning                                                  |
| --------- | -------------------------------------------------------- |
| `output`  | The directory, inside the site, that is published        |
| `command` | A list of words, like the preview command                |

Without `build`, a site that declares a runtime has nothing to publish by a
build.

## Placeholders

A word that is exactly one of these is replaced by ChatCMS. When it has no
value, the word and a flag just before it are left out.

| Placeholder | Value                                               |
| ----------- | --------------------------------------------------- |
| `{host}`    | The address to listen on                            |
| `{port}`    | The port to listen on                               |
| `{url}`     | The address the preview is reached at               |
| `{reload}`  | The port for live reload, where there is one        |
| `{source}`  | The site's directory                                |
| `{output}`  | Where a build writes what is published              |
| `{cache}`   | A directory to keep a cache in                      |

Any other word in braces is refused.

## When it changes

A changed definition takes effect the next time the preview is opened. One
that cannot be read, or that names a program no declared package provides,
stops the preview and says why to implementers.

## Outputs

A publish workflow may say what of the build goes where:

```yaml
on: publish
steps:
  - action: build
  - action: publish
outputs:
  - path: .
    to: publish
  - path: "*.pdf"
    to: both
```

A path is a file, a directory, or a pattern, inside the build's output; `.`
is all of it. `publish` sends it with the site, `artifact` keeps it in
ChatCMS, `both` does both. Without `outputs` the whole build is published.
Kept files are listed under Artifacts and offered on the publish's Done
screen; each publish keeps its own, up to 200 MB a file.
