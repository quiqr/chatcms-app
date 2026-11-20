## 1. Implementation

- [x] 1.1 A page says what its preview shows (own, declared, none, hidden),
  from the model's `previewUrl` and `hidePreviewIcon` and the file's kind;
  the preview shows the address, the explanation with "Show the home page",
  or that it is switched off. Bean chatcms-a6on.

## 2. Verification

- [x] 2.1 Go test `TestLocateSaysWhatThePreviewShows`; Vitest for the
  explanation, the home button, a declared and a hidden preview; e2e
  `52-no-page-of-its-own` at both widths.
