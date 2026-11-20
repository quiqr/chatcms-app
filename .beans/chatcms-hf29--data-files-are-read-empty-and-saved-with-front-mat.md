---
# chatcms-hf29
title: Data files are read empty and saved with front matter
status: completed
type: bug
priority: high
created_at: 2026-11-20T14:21:57Z
updated_at: 2026-11-20T16:04:12Z
openspec-link: openspec/changes/archive/2026-11-20-data-files-whole
---

Found while building chatcms-a6on. A single or collection item whose file is
data rather than a page is handled as front matter:

- `data/hours.yaml` and `config.toml` have no `---` or `+++` markers, so
  `frontmatter.Parse` reads them as having none: every field in the form is
  empty.
- Saving such a file writes `---` YAML front matter before the old content,
  which breaks the file (for `config.toml` the whole site).
- A JSON data file is read, but saving it is refused
  (`ErrUnwritableFormat`).

Seen with UILite from the gallery (`config.toml` as Site Configuration,
`data/*.json`). Probed on 20 November 2026 with `forms.Load`:
yaml read empty / written with front matter, toml the same, json read but not
written.

Expected: a JSON, YAML or TOML file named by the model is the data as a whole,
read and written in its own format, keeping what the form does not touch; old
Quiqr's output is the golden for byte compatibility.

## Summary of Changes

- A file that is not `.md`, `.markdown`, `.html` or `.qmd` is read and written
  whole by its extension, as Quiqr desktop's format providers do.
- JSON is written as `JSON.stringify(value, null, 2)` and TOML as tomlify-j0.4
  with `space: 2`, both without a final newline; proven byte for byte against
  all 156 recorded cases. YAML as js-yaml, as before.
- TOML is read with its key order kept at every depth, and arrays of tables
  as lists.
- e2e 53 edits `data/hours.json` and `hugo.toml`, publishes, and finds the
  reference's bytes on the remote.
