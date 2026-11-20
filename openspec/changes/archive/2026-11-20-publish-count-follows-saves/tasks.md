## 1. Implementation

- [x] 1.1 A successful write to a project announces it (`tellProjectChanged`
  in `lib/api.ts`), as does an assistant's change; the screens that hold the
  project read it again at once and two seconds later, and when the tab
  becomes visible (`useProjectChanges`). Bean chatcms-9m9u.

## 2. Verification

- [x] 2.1 Vitest `projectchanges.test.tsx`; e2e
  `50-the-publish-count-follows` passes at both widths.
