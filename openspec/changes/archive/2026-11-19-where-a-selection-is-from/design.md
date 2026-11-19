# Design

## Context

`assistant.Web.Ask` turns a selection into a prompt. Its only knowledge of
where the text comes from is `pageLocator`, which runs `hugo list all` in the
sandbox and maps the preview address to a content file, so only Hugo sites
get even the page. The content package already reads any page as a form
(`frontmatter.Parse`, `forms.New`), and the model lists every single and,
through `model.Items`, every item of a collection.

## Decisions

### 1. Search the source, not the HTML

`anchor.Find(site, model, asked)` reads the pages the model names and
compares the selection with their field values and bodies as plain text.
This works whatever turned the source into HTML. A generator-specific step
may narrow it (the page from Hugo's list) but is never needed.

### 2. Order

The hinted page first; then shared singles (file outside `content/`); then
every other page. No match is the design. Within the matches, the field
whose plain text also holds the last words before the selection or the first
words after it wins; otherwise the first in model order.

### 3. Plain text

Shortcodes, link definitions, link and image syntax (keeping the words),
HTML tags, heading, quote and list markers and emphasis marks are removed,
entities decoded, whitespace collapsed, case folded. The selection from the
browser gets the same treatment.

### 4. Data files

A shared single whose file has no front matter and is YAML, JSON or TOML is
read line by line as `key: value`, which is enough to say which key holds
the text.

### 5. Where it runs

`app.sitePlacer` reads the person's own copy of the site (the same copy the
assistant edits) and is set on the assistant's web layer independently of
the sandbox, so it also works where Hugo is not run.

## Risks / Trade-offs

- **Cost**: every request reads the pages; bounded at 1000 pages, and a
  site's text is small. Caching can follow when it shows.
- **Text built by templates** from several fields (a date plus a title) is
  found by neither and reads as the design; tier 1 markers will say better.
