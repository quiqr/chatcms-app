# Proposal

Bean: [chatcms-nwic](../../../.beans/chatcms-nwic--a-file-manager-for-templates-and-layouts.md)
(a file manager for templates and layouts).

## Why

Implementers can change a site's model and content through ChatCMS, but its
templates, layouts, styles and configuration only from outside. They should
be able to look through a project's files and change a template the way
they change anything else: in their own copy, saved and published like the
rest.

## What Changes

- **Setup → Files**, for whoever may see a project's technical detail: the
  directory tree of their own copy of the site, as an accessible tree
  (`role="tree"`, arrow keys, one tab stop), and a click on a file opens it
  in a code editor beside it (on a phone, filling the screen).
- **The editor** is CodeMirror 6, loaded only on this screen, with line
  numbers, search, indenting, and highlighting for HTML (with Go template
  actions `{{ }}` marked), CSS, JavaScript, JSON, YAML, TOML and Markdown.
- **Saving** writes the file into the person's own copy as a saved change,
  which reaches the shared version and is published like every other change.
  A file that changed since it was opened is not overwritten.
- **The server** lists the tree and reads and writes one text file at a time,
  with a path that cannot leave the project (`..`, absolute paths, symbolic
  links) and never touches `.git`.

## Capabilities

### New Capabilities

- `content/project-files`: browsing, reading and editing a project's files.

### Modified Capabilities

## Impact

- `internal/content`: the file service; `internal/httpapi`: three routes.
- `web/src`: the Files page, the tree, the lazily loaded editor; strings.
- `web/package.json`: CodeMirror packages, pinned to the versions Milkdown
  already brings, so there is one copy of `@codemirror/state`.

## Assumptions

- Agreed on 20 November 2026: a hand-made tree with the ARIA tree pattern,
  CodeMirror 6 with an own wrapper, and an own overlay for Go templates;
  not Monaco, not react-arborist, not a CodeMirror wrapper package.
- Text files up to 1 MB are opened; a larger or binary file says so.
- Only existing files are edited; making, renaming and removing files come
  later.
- `public/`, `resources/_gen/` and `node_modules/` are build output and
  left out of the tree.
