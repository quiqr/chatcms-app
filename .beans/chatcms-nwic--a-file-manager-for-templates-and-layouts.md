---
# chatcms-nwic
title: A file manager for templates and layouts
status: in-progress
type: feature
priority: normal
created_at: 2026-11-20T13:34:53Z
updated_at: 2026-11-20T18:58:33Z
---

Implementers can look through a project's files and read them: the
directory tree, and a click on a file shows it with syntax highlighting.
Editing templates and layouts comes with it, as a plain editor rather than
the form-based one content creators see.

Chosen on 20 November 2026, after comparing components:

- The tree by hand: a recursive list with `@radix-ui/react-collapsible`
  (already a dependency) and a lucide icon per kind. `react-arborist` only
  if a project ever holds thousands of files.
- The editor: CodeMirror 6, loaded lazily (`React.lazy`) so no other screen
  pays for it, with an own theme over the CSS variables in `index.css`.
  Languages: `lang-yaml`, `lang-html`, `lang-css`, `lang-javascript`,
  `lang-markdown`, `codemirror-lang-toml`. Hugo templates have no Go
  template mode: `lang-html` leaves `{{ }}` unhighlighted, which is
  accepted for a simple editor.
- Not Monaco (about 5 MB, workers, a Vite plugin) and not a textarea with
  Shiki or lowlight (no line numbers, indenting or searching).

The real work is on the server: an endpoint for the tree and one for a file,
with a path that cannot leave the project (`..`, symlinks) and with the
same permission as the rest of the project's technical detail.
