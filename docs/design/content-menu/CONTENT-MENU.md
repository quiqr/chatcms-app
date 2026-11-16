# Quiqr ChatCMS — Content menu & collections (direction B)

Spec for replacing the current full-screen content overview with a **content
column driven by the Quiqr model menu**, plus **collection views** (lists of
items such as blog posts) and a **content-aware editor header**.

Reference screens: `reference/html/*.html` (open in a browser; exact sizes and
colours are in the inline styles) and `reference/screenshots/{desktop,phone}`.
Visual system: same tokens and components as the milestone-1 design
(`tokens/tokens.css`, `DESIGN.md` from the main design kit). Where this file
and an earlier spec disagree about the content area, this file wins.

Sample data in the references (groups, page names, posts, people) is made up
and only illustrates structure.

---

## 1. Concepts

| Quiqr model | ChatCMS UI | Notes |
| --- | --- | --- |
| `menu` (groups with menu items) | Content column groups, in model order | Group title = menu group title. |
| Menu item pointing to a **single** | Row with page icon → opens the form editor | |
| Menu item pointing to a **collection** | Row with stack icon + item count → opens the collection view | |
| Collection subfolders | Folder tabs in the collection view | e.g. language folders `nl/`, `en/`. |
| Collection sort key / item title field | Default sort + row title | Fall back to a humanised file name when no title field exists. |
| Hugo `draft` field (when present in the model) | "Show on website" switch in the editor, "Draft" badge in lists | Never show the word `draft:` or the field name to editors. |

**No `menu` in the model?** Build two groups automatically: *Pages* (all
singles, A–Z) and *Collections* (all collections, A–Z).

## 2. Layout

```
Desktop, collection or overview open      Desktop, item open (editor)
┌────────┬──────────────┬───────────────┐ ┌──┬──────────────┬────────────┬──────────┐
│ app    │ content      │ collection    │ │ic│ content      │ form       │ preview  │
│ sidebar│ column       │ view          │ │on│ column       │            │          │
│ 220–240│ 250–260      │ fluid         │ │64│ 250 (fold)   │ fluid      │ fluid    │
└────────┴──────────────┴───────────────┘ └──┴──────────────┴────────────┴──────────┘
```

- **Collection / overview open** (`content`): full app sidebar (Content
  active) + content column + main area.
- **Item open** (`editor`): the app sidebar collapses to the 64 px icon rail
  (project tile, Content, Media, Publish, History, avatar). The content column
  stays. Its header has a **fold** button that hides it (`editor-menu-hidden`);
  the Content icon in the rail brings it back. Remember the fold state per user
  (local storage is fine).
- **Below 1024 px**: the content column becomes an overlay drawer opened from
  the Content icon; it closes when an item is chosen.
- **Phone (< 640 px)**: no columns. Content tab → grouped menu screen
  (`phone-content-menu`) → collection screen (`phone-collection`) → existing
  phone editor. Bottom tab bar: Content, Media, Publish (badge), History.

## 3. Content column

- Header: "Content" + summary ("35 pages · 2 collections").
- **Quick find** field, also opened with **⌘K / Ctrl+K** from anywhere in the
  project: searches titles of all singles and all collection items; results
  grouped by group/collection; Enter opens.
- **Groups** in model order. Group header = uppercase label, chevron,
  collapse/expand on click. A collapsed group shows its item count. Remember
  collapsed state per user per project.
- **Rows** (height 34–36 px, radius 8):
  - Single: page icon + title.
  - Collection: stack icon + title + item count (right, muted).
  - Orange dot (`warning`) when the item, or anything in the collection, has
    unpublished changes.
  - Active row: `primary-soft` background, `primary-soft-foreground` text,
    600 weight. **A collection stays active while one of its items is open.**
- Long titles truncate with an ellipsis; full title in a tooltip.

## 4. Collection view (`content`, `content-no-results`, `content-new-post`)

- Top bar: breadcrumb `Content / <group> / <collection>`, sync status, Publish.
- Heading: collection title (26 px / 650) + line "48 posts · newest first"
  (the noun comes from the collection's item label: post, case, …).
- Primary action: **New <item>** (ink button).
- Toolbar (wraps on narrow screens):
  - Search within this collection (title + summary), 250 ms debounce.
  - **Folder tabs** (segmented control): "All · n" + one tab per top-level
    subfolder with counts. Hide the control when there are no subfolders. More
    than 5 folders → use a select instead.
  - Sort select: Newest first (default, or the model's sort key), Oldest
    first, Title A–Z, Recently edited.
  - **File names** toggle — implementers, organisation admins and platform
    admins only. Shows the repository path under each title (mono, 12 px).
- List (card with rows):
  - Thumbnail 56×42 when the model has an image field; omit otherwise.
  - Title (15 px / 500) → opens the item.
  - Meta line: date · author · folder/language (whatever exists in the model).
  - Status badge: Changed (warning), New (info), Draft (muted), Published
    (success). Hide "Published" on phones to reduce noise.
  - Row menu: Open, Rename, Duplicate, Move to folder…, Delete….
  - Footer: "Showing n of N" + "Show 20 more" (page size 20).
- **No results**: card with "No posts match “<query>”", one helpful sentence,
  **Clear search**.
- **Empty collection**: "No posts yet" + New post button.
- **New item dialog**: Title (required), Folder/Language (only with
  subfolders), web address preview with Edit (slug from title), and for
  implementers the file path that will be created. Buttons: Cancel /
  **Create and edit** (opens the editor). If the slug or file already exists,
  show "There’s already a post at this address. Change the title or the
  address." and keep the button disabled.

## 5. Editor header when an item is open (`editor`)

- Back button → the collection (or the group's first view for singles).
- Context line above the title: "<collection> · 12 of 48" (link to the
  collection); for singles: the group name.
- **Previous / next** buttons step through the collection in the current
  sort, search and folder filter (disabled at the ends). Keyboard: Alt+↑ / Alt+↓.
- Then the existing editor controls: Saved status, Form / Side by side /
  Preview, Publish.
- "Show on website" switch at the top of the form when the model has a
  `draft` field (inverted: on = not a draft).

## 6. Roles

Everyone who can read the project sees the menu and collections. Viewers get
no New / Rename / Duplicate / Move / Delete. File names toggle and file paths
only for implementers and admins (see DESIGN.md §6).

## 7. Suggested API shape (adjust to the backend)

```
GET  /api/projects/{id}/content-menu
     → { groups: [{ key, title, items: [{ key, kind: "single"|"collection",
          title, count?, hasChanges }] }], totals: { singles, collections } }
GET  /api/projects/{id}/collections/{key}/items
       ?q=&folder=&sort=date_desc|date_asc|title|edited&page=&pageSize=20
     → { items: [{ id, title, date, author, folder, status, thumbnail?, path }],
         total, folders: [{ name, label, count }] }
POST /api/projects/{id}/collections/{key}/items  { title, folder?, slug? }
     → 201 { id } | 409 { error: "slug_taken" }
GET  /api/projects/{id}/search?q=   (quick find across singles + items)
```

`status` is derived server-side from the working copy vs. the last publish
(+ the draft field). `path` is only returned to implementers/admins.

## 8. Acceptance checks

1. The content column lists the model's menu groups and items in model order;
   a model without `menu` shows the Pages/Collections fallback.
2. Clicking a collection shows its items with titles (not file names), counts,
   status badges and working search, folder tabs and sort.
3. Opening an item keeps its collection highlighted, shows "n of N" and
   previous/next follow the current list order and filters.
4. The content column folds and unfolds; the choice survives a reload.
5. ⌘K finds singles and collection items by title.
6. New item: a duplicate address is refused with the message from §4; a valid
   one opens the editor on the new item.
7. Editors never see file paths; implementers can toggle them.
8. Phone: Content tab → menu → collection → editor works at 390 px with no
   horizontal scrolling and 44 px touch targets.
