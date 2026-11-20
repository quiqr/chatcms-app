## 1. Implementation

- [x] 1.1 The preview reloads 2.5 and 8 seconds after an assistant's change
  (`settle`), besides at once. Found while shipping chatcms-hf29: e2e 37
  failed under load because the one reload came before Hugo had rebuilt.

## 2. Verification

- [x] 2.1 Vitest for the two later reloads; e2e 37, 35, 39 and 47 five times
  each at 12 workers: 50 of 50 passed, where without it 37 failed 1 to 3 of
  10 in four runs.
