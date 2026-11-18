# Quiqr ChatCMS — folded content menu: vertical strip + breadcrumb picker

Small change to the existing content column. Nothing else in the content
area changes.

Today, when the content column is folded, only an unfold icon remains: you
can't see what's behind it or where you are. This adds two things that work
together:

1. a **vertical label strip** that replaces the bare icon;
2. a **location picker** in the breadcrumb, to switch without unfolding.

References: `reference/screenshots/content-menu-folded.png` (strip, picker
closed) and `content-menu-folded-picker.png` (picker open). The same states as
static HTML with exact inline styles are in `reference/html/`. Use the existing
tokens (`tokens/tokens.css`) and shadcn components, not inline hex values.

---

## 1. Vertical label strip (48 px)

Replaces the current folded state of the content column (currently ~72 px
wide, icon only).

- Width **48 px**, full height, background `#FBFBFA`, right border
  `border` (same as the unfolded column).
- The **whole strip is one button**, `aria-label="Show content menu (<current
  item>)"`. Clicking anywhere on it unfolds the column. Hover: `muted`
  background on the whole strip; visible focus ring.
- Contents, top to bottom, centred, gap 16 px, padding-top 14 px:
  1. The unfold icon (the existing panel icon) in a 36 px square with `muted`
     background, radius 8.
  2. A vertical label (`writing-mode: vertical-rl`, reads top to bottom, gap
     10 px between parts):
     - `CONTENT` — 12 px, 600, uppercase, letter-spacing .08em,
       `muted-foreground`;
     - `·` separator in `#A1A1AA`;
     - the **current item's title** — 13 px, 600, `primary-soft-foreground`;
     - for a collection, its **item count** — 13 px, `subtle-foreground`.
     On the content home (nothing selected) show only `CONTENT`.
  3. An orange `warning` dot (7 px) when anything in the content menu has
     unpublished changes (same rule as the dots in the unfolded menu).
- Long titles: truncate with an ellipsis so the label never takes more than the
  strip height minus 120 px; full title in a tooltip.
- The fold state stays remembered per user as it is now.

## 2. Location picker in the breadcrumb

Page header of content screens (collection view and editor): breadcrumb
`Content / <group> / <current item ▾>`.

- The **last breadcrumb part is a button** (`aria-haspopup="menu"`,
  `aria-expanded`). At rest: plain text (600) with a chevron, no border. Open:
  1 px `primary` border, `primary-tint` background, `primary-soft-foreground`
  text, radius 8, height 34.
- Available whether the column is folded or not.
- **Popover**: 320 px wide, anchored under the button, radius 12, menu shadow,
  padding 8:
  - **Quick find** field with the `Ctrl K` / `⌘K` hint — the same search as
    the content column's quick find.
  - The model menu in compact form, in model order:
    - group labels 11 px, 600, uppercase, `subtle-foreground`, with the item
      count on the right;
    - items 32 px high, 14 px; collections show their count on the right;
    - the current item highlighted (`primary-soft` background,
      `primary-soft-foreground`, 600);
    - groups with **more than 6 items** show only label + count; clicking the
      label expands that group inside the popover.
  - Footer: divider + text button **Show the full content menu**, which
    unfolds the column and closes the popover.
- Choosing an item navigates to it and closes the popover.
- Keyboard: Enter/Space opens; arrow keys move through items; Enter opens the
  focused item; Esc closes and returns focus to the button. Typing goes into
  quick find.
- Long popovers scroll inside (max height: viewport minus 120 px).
- Phones (< 640 px): no picker; the existing phone flow stays.

## 3. Acceptance checks

1. Folded, the strip is 48 px wide and shows `CONTENT · <current item>
   <count>` vertically; on the content home only `CONTENT`.
2. Clicking anywhere on the strip unfolds the column; the fold state survives
   a reload.
3. The change dot appears on the strip exactly when the unfolded menu would
   show at least one dot.
4. A long title is truncated in the strip and shown in full in the tooltip.
5. The breadcrumb's last part opens the picker, folded or not; picking an
   item navigates and closes it; Esc closes and returns focus.
6. Quick find in the picker finds the same results as the column's quick find.
7. "Show the full content menu" unfolds the column.
8. Both work with keyboard only and pass a basic screen-reader check (button
   labels, `aria-expanded`).
