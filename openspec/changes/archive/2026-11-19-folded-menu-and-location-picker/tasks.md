# Tasks

## 1. The strip

- [x] 1.1 The folded column as one 48 px button with the vertical label, the
  count and the dot; a tooltip with the whole title; proven by Vitest
  (checks 1, 3, 4 of FOLDED-MENU.md section 3).
- [x] 1.2 Unfolding from elsewhere by an event; the strip unfolds and the fold
  survives a reload; proven by Vitest (check 2).

## 2. The picker

- [x] 2.1 `components/ui/popover.tsx`.
- [x] 2.2 `ContentBreadcrumb` with the picker: quick find, the compact menu,
  large groups closed, choosing navigates and closes, the footer unfolds;
  proven by Vitest (checks 5, 6, 7).
- [x] 2.3 Keyboard: arrows, Enter, Esc with focus back; labels and
  `aria-expanded`; proven by Vitest (check 8).
- [x] 2.4 The breadcrumb on the collection view and in the editor, with the
  position kept; proven by Vitest and e2e `43-the-folded-menu` at 1440 px.

## 3. Ship

- [x] 3.1 Screenshots of both states next to the references.
- [x] 3.2 `CHANGELOG.md` under [Unreleased].
- [x] 3.3 `scripts/gate.sh ship` green; close chatcms-vqsf.
