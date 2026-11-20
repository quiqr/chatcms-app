## 1. Implementation

- [x] 1.1 Remove `SyncIndicator` from EditorShell, ProjectShell and
  PhoneShell, and delete `shells/SyncIndicator.tsx`, its `syncState` prop and
  its strings.
- [x] 1.2 The page editor's status says "Saved" (nl "Opgeslagen") under a
  `page-editor.*` key; "Saving…" comes from the catalogue too; drop the
  comment about two live regions.
- [x] 1.3 Tests that looked for "All changes saved" look for "Saved"; the
  e2e `saves` helper too. Bean chatcms-gpmw.

## 2. Verification

- [x] 2.1 Typecheck, `check:strings`, Vitest, and the e2e specs that save
  (03, 09, 32, 33, 50) at both widths: 479 Vitest and 17 e2e passed.
