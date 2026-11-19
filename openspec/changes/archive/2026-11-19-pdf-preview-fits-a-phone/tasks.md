<!-- Schema: https://github.com/speclib/openspec-tinychange-schema -->

## 1. Implementation

- [x] 1.1 The pane asks the PDF viewer to fit the width on a phone
  (`&fit=width`); the viewer scales the pages down to the room, never up.
- [x] 1.2 The development seed previews the recipes book as its PDF (Typst);
  the e2e fixture keeps its website preview, which e2e 38 tests as well.

## 2. Verification

- [x] 2.1 Vitest for the scale and the address; Go test for the seed's
  runtime, including against the fixture book.
- [x] 2.2 e2e 38 "previewed as a PDF" at both widths: own size on the
  desktop, within the screen on a phone.
