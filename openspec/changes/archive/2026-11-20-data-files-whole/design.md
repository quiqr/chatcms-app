# Design

## Context

`frontmatter.Parse` splits delimited front matter from a body and is used for
every file. A file without delimiters is "no front matter": all of it is
body. `Document.Bytes` writes YAML front matter only. Old Quiqr decides by
extension (`isContentFile`): content files go through front matter, anything
else is parsed and dumped whole by its format provider.

## Decisions

- **`Document.Whole`** marks a document that is the whole file. `ParseData`
  reads one; `Bytes` writes it in its own format. `forms.Load` picks by
  extension with the reference's list of content extensions.
- **JSON emitter** follows `JSON.stringify` with two spaces: keys in emit
  order (array-index keys first, as JavaScript does), numbers as JavaScript
  writes them, dates as ISO strings with milliseconds, no HTML escaping.
- **TOML emitter** is a port of tomlify-j0.4: plain values first, then tables
  and arrays of tables, nested tables indented two spaces per level, a blank
  line before each table header, arrays over several lines, every number as a
  float, keys bare when they match `[A-Za-z0-9_-]+`, null values left out.
- **TOML reading keeps order** by walking the decoder's key list for every
  path; an array of tables becomes a list of maps.
- **Goldens**: the existing `.json.golden` (compared without its final
  newline) and `.toml.golden` files are the reference.

## Risks / Trade-offs

- A hand-written data file is rewritten in the reference's style on its first
  save: comments go, integers become floats in TOML. That is what old Quiqr
  does, and adopting ChatCMS should not differ from it.
