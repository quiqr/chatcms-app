## 1. Implementation

- [x] 1.1 Signing out stops a refused request from sending the person to the
  sign-in screen; signing in again restores it. Bean chatcms-0xlf.
- [x] 1.2 The overlay reports an empty page with `ready`; the preview loads
  it again (1.5 s, 3 s, 4.5 s, at most three times); the PDF viewer never
  counts as empty. Bean chatcms-v58o.
- [x] 1.3 The page editor's quiet time before saving can be shortened, and
  the undo tests use 20 ms instead of waiting 800 ms four times on a busy
  machine. Bean chatcms-1jm8.

## 2. Verification

- [x] 2.1 Vitest for each (499 and more pass); e2e 14, 33, 37, 05 and 38
  four times each at 12 workers: 14, 33 and 37 passed every time, where they
  failed before; 38's Quarto start failed twice in one run and passed in the
  next, and is left for its own bean.
