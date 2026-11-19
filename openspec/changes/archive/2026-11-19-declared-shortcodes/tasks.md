# Tasks

## 1. The model

- [x] 1.1 `shortcodes` in base and includes, merged and checked; fixture
  `test/fixtures/models/shortcodes` with goldens; proven by Go tests.
- [x] 1.2 `shortcodeDialect` in the base, checked against the dialects this
  version has; the page read returns the resolved dialect (model, else
  runtime) and the declarations as controls without values; proven by
  model and httpapi tests.
- [x] 1.3 The editor loads the dialect the page names, or none; a Quarto
  page keeps its shortcode text as plain text; proven by Vitest.

## 2. Parameters

- [x] 2.1 The tokenizer: names, positional and named parameters, quotes,
  backticks, inner text; refuses what it cannot read; proven by Vitest.
- [x] 2.2 Rewriting: a changed, a new and a cleared value, the inner text,
  everything else byte for byte; building a new tag; proven by Vitest.

## 3. The editor

- [x] 3.1 The node view for declared shortcodes: title, first text value,
  thumbnail; undeclared ones stay locked; proven by Vitest.
- [x] 3.2 The form dialog with `Fields` and the inner text box; Done as one
  transaction; Cancel changes nothing; proven by Vitest.
- [x] 3.3 The Insert menu in the toolbar; absent without declarations.
- [x] 3.4 Strings in `en.ts` and `nl.ts`.

## 4. Proof and ship

- [x] 4.1 e2e: a site declaring `figure` and `notice`; change a caption,
  pick a picture, change a notice's text, insert a figure; the saved file
  differs only there; at both widths.
- [x] 4.2 `CHANGELOG.md` under [Unreleased].
- [x] 4.3 `scripts/gate.sh ship` green; close chatcms-oren.
