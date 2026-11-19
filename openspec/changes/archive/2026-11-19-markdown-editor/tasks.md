# Tasks

## 1. Writing back faithfully

- [x] 1.1 `@milkdown/kit` added; `npmDepsHash` regenerated.
- [x] 1.2 The source keeper: blocks with ranges, matching on `node.eq`,
  splicing; empty bodies; proven by Vitest on the fixture corpus and
  `test/fixtures/markdown/tricky.md` (untouched, edit, insert, remove).
- [x] 1.3 The mismatch check that sends text to the Markdown view; proven by
  Vitest.

## 2. The Hugo dialect

- [x] 2.1 The remark transformer for block, paired and inline shortcodes, and
  the two locked nodes; written back verbatim; proven by Vitest.

## 3. The editor

- [x] 3.1 The Milkdown editor module: label, role, save only on a person's
  edit, rebuild on an outside value; proven by Vitest.
- [x] 3.2 The toolbar on editor commands, link and image by popover, labels
  from the catalogue.
- [x] 3.3 `BodyEditor`: formatted and Markdown views, lazy loading, the
  Markdown-only note; existing editor tests adapted.

## 4. Proof and ship

- [x] 4.1 e2e: bold a word in the formatted view, the saved file has it in
  Markdown and is otherwise unchanged; a shortcode survives an edit; at both
  widths.
- [x] 4.2 `CHANGELOG.md` under [Unreleased].
- [x] 4.3 `scripts/gate.sh ship` green; keep chatcms-oren open for the next
  change.
