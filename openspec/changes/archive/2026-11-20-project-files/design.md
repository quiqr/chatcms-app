# Design

## Context

The content service opens a person's own copy (`Sources.Site`) as an
`fs.FS` and saves through `Site.Save(gitsync.Save{Writes})`, which makes a
version, after which `SyncSoon` brings it to the shared copy. Permission for
technical detail is `policy.ActionSeeTechnicalDetail`.

## Decisions

- **In `internal/content`**, beside pages: `Files` (the tree), `ReadFile` and
  `WriteFile`, through the same `Sources` and `Site`, so a saved file is a
  version exactly like a saved page.
- **Paths**: `fs.ValidPath` rules out `..` and absolute paths; every
  component is checked with `fs.Lstat` so a symbolic link anywhere on the way
  is refused; `.git` is refused by name. The tree does not follow links and
  shows them as files that cannot be opened.
- **Not overwriting**: reading answers a `sum` (SHA-256 of the bytes);
  writing sends the sum it read, and a different current sum is refused with
  409 and `refused.file-changed`.
- **Limits**: at most 5000 entries in the tree; a file over 1 MB or with a
  NUL byte or invalid UTF-8 is answered as not openable.
- **Routes**: `GET …/files`, `GET …/files/{path...}`, `PUT …/files/{path...}`.
- **Frontend**: `FilesPage` with a tree component that keeps the focused
  item and the open folders, roving tabindex, and the keys of the ARIA tree
  pattern. The editor is a separate module behind `React.lazy`: a small
  wrapper around `EditorView` with `basicSetup`-like extensions chosen by
  hand, a theme over the CSS variables, the language by extension, and a
  `MatchDecorator` overlay for `{{ … }}`.
- **Packages**: `@codemirror/state`, `view`, `language` pinned to Milkdown's
  versions; `commands`, `search`, `lang-html`, `lang-css`,
  `lang-javascript`, `lang-json`, `lang-yaml`, `lang-markdown` and
  `legacy-modes` (for TOML).

## Risks / Trade-offs

- A template saved with a mistake breaks the preview; the preview says so and
  the History screen can go back. The editor does not validate templates.
