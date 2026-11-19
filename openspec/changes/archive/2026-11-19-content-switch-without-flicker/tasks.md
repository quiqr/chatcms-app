<!-- Schema: https://github.com/speclib/openspec-tinychange-schema -->
<!-- Bean: .beans/chatcms-xlko--switch-content-element-reloads-page.md -->

## 1. Implementation

- [x] 1.1 `useAsync` takes an optional name under which it keeps its last
  answer, starts from it, and loads again behind it; `signOut` forgets them.
- [x] 1.2 The project and the content menu are kept by name on the content
  screens and in the editor.

## 2. Verification

- [x] 2.1 Vitest for `useAsync` (kept, unkept, forgotten, changing names) and
  the whole Vitest suite.
- [x] 2.2 e2e 43 records whether the frame disappears while switching: it does
  against the build before this change and does not after it.
