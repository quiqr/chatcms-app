# The runtime definition

A site may say how it is previewed and built in `quiqr/runtime.yaml`. A site
without one is previewed and built by Hugo at the project's recorded version,
as before.

```yaml
packages:
  quarto: "1.6.40"
  hugo: "0.139.0"
preview:
  type: web
  open: /
  command: [quarto, preview, "{source}", --host, "{host}", --port, "{port}", --no-browser]
build:
  output: _site
  command: [quarto, render, "{source}", --output-dir, "{output}"]
```

## packages

Each name and version is looked up on Nixhub and built with Nix, once per
version, on installations with Nix (`CHATCMS_NIX_BIN`). Their `bin`
directories make up the command's `PATH`. Without Nix only `hugo` can be
named, and the installation's Hugo is used.

## preview

| Field     | Meaning                                                        |
| --------- | -------------------------------------------------------------- |
| `type`    | `web` (the default), `pdf` or `other`                          |
| `open`    | The page the pane opens, starting with `/`; `/` by default     |
| `command` | A list of words; the first is a program from a declared package |

A `web` preview follows the page being edited. A `pdf` preview shows `open`
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
