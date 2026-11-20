---
# chatcms-nwic
title: A file manager for templates and layouts
status: completed
type: feature
priority: normal
created_at: 2026-11-20T13:34:53Z
updated_at: 2026-11-20T19:20:27Z
openspec-link: openspec/changes/archive/2026-11-20-project-files
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

## Summary of Changes

- Setup → Files for implementers: the tree of their own copy (without `.git`
  and build output) with the ARIA tree pattern, and a CodeMirror 6 editor
  loaded only on this screen (a 44 KB chunk).
- Languages by extension; Go template actions `{{ }}` marked by an own
  overlay; a theme over the CSS variables; Ctrl/Cmd+S saves.
- Saving writes into the person's copy as a saved change, published like any
  other; a file that changed since it was opened is not overwritten.
- The server refuses `..`, absolute paths, symbolic links and `.git`; the
  content FS gained `Lstat` and `ReadLink` so links are seen.
- CodeMirror's core packages are pinned to Milkdown's versions: one copy of
  `@codemirror/state`.
