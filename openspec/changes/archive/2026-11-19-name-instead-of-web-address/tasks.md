<!-- Schema: https://github.com/speclib/openspec-tinychange-schema -->
<!-- Bean: .beans/chatcms-gfcb--web-address-is-not-relevant-for-pdfs.md -->

## 1. Implementation

- [x] 1.1 The server gives a new item a web address only on a Hugo site
  (`RuntimeOf`), shared with the shortcode dialect.
- [x] 1.2 The new-item dialog shows "Name" and the slug where no web address
  comes.

## 2. Verification

- [x] 2.1 Go test: a Hugo site gets `/news/spring/`, a Quarto site only the
  name; Vitest: the dialog says Web address or Name.
