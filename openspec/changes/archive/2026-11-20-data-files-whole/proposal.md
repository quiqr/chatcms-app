# Proposal

Bean: [chatcms-hf29](../../../.beans/chatcms-hf29--data-files-are-read-empty-and-saved-with-front-mat.md)
(data files are read empty and saved with front matter).

## Why

A single or collection item whose file is data rather than a page, such as
`config.toml` or `data/services.json` in the UILite template, is handled as a
content file with front matter. YAML and TOML data is read as having no
values, saving it puts YAML front matter in front of the old content and
breaks the file, and saving JSON data is refused. Old Quiqr reads and writes
such a file whole, in its own format.

## What Changes

- A file whose extension is not `.md`, `.markdown`, `.html` or `.qmd` is
  read whole as JSON, YAML or TOML by its extension, as old Quiqr does.
- It is written whole in the same format, byte for byte as the reference:
  JSON as `JSON.stringify(value, null, "  ")` without a final newline, YAML
  as js-yaml writes it, TOML as tomlify-j0.4 with `space: 2` without a
  final newline.
- TOML is read keeping the order of keys at every depth, and an array of
  tables is read as a list, so writing back changes only what was edited.
- Front matter in TOML or JSON stays read-only, as decided before.

## Capabilities

### New Capabilities

### Modified Capabilities

- `content/serialisation`: data files are read and written whole, matching
  the reference.

## Impact

- `internal/frontmatter`: whole-file documents; JSON and TOML emitters;
  ordered TOML reading.
- `internal/forms`: chooses whole-file or front matter by extension.
- Goldens: the existing `.json.golden` and `.toml.golden` files become
  tested; no new reference run is needed.

## Assumptions

- The reference writes whole numbers in TOML as `3.0`; ChatCMS does the same,
  so a hand-written `paginate = 10` becomes `paginate = 10.0` on the first
  save, as it would in old Quiqr. Hugo reads both.
- Comments in a hand-written data file are lost on save, as in old Quiqr.
- The JSON goldens were recorded with a final newline the application does
  not write; the test compares without it.
