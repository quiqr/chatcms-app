# Prompt for Claude Code

Put this folder in the frontend repo as `docs/design/content-menu/` (next to
the main design kit, if you have it), then paste:

```
We're replacing the full-screen content overview in Quiqr ChatCMS with the
content-menu design in docs/design/content-menu/.

Read docs/design/content-menu/CONTENT-MENU.md fully first. Reference screens
are in docs/design/content-menu/reference/screenshots/ (desktop 1440 px,
phone 390 px) and reference/html/ (static HTML with exact inline styles).
Use our existing tokens and shadcn components, not the inline hex values.

Work in this order and stop after each step to show me screenshots at
1440 and 390 px next to the matching reference:
1. Backend: content-menu endpoint from §7, built from the Quiqr model `menu`
   (with the Pages/Collections fallback), including counts and hasChanges.
2. Content column (§3) inside the project layout, with collapse state and
   ⌘K quick find.
3. Collection view (§4): list, search, folder tabs, sort, pagination,
   no-results and empty states, File names toggle for implementers.
4. New item dialog (§4) with the slug/address check.
5. Editor header (§5): back, "n of N", previous/next, foldable content
   column with the 64 px icon rail, "Show on website" draft switch.
6. Phone flow (§2): menu screen → collection screen → editor.

Finish by running the acceptance checks in §8 and listing any that fail.
Don't show git terms or file paths to editors.
```
