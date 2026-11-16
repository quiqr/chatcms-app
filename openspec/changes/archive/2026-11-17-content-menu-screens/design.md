# Design

## Context

After `content-follows-the-model`, the server answers the content menu
(`GET …/content-menu`) and a collection's items
(`GET …/collections/{collection}/items`) by Quiqr Desktop's rules, and the
current `ContentPage.tsx` shows the menu as one long page. The editor
(`EditorPage.tsx`) already has a 64 px rail and an empty slot, which
`interface/app-shell` kept for this. The phone frame already has a bottom tab
bar. The hand-off is `docs/design/content-menu/CONTENT-MENU.md` with reference
HTML and screenshots at 1440 and 390 px.

## Goals / Non-Goals

**Goals:**

- The screens of direction B at 1440, 1024 and 390 px, from our tokens and
  shadcn components rather than the reference's inline values.
- Every list a person scrolls is filtered, sorted and paged by the server, so a
  collection of a thousand items is as quick as one of ten.

**Non-Goals:**

- Moving an item to another folder (see the proposal's assumptions).
- Searching inside page text. Quick find is titles; the collection search is
  title and summary.

## Decisions

### One content layout, three routes

A `ContentLayout` holds the content column and an outlet:

| Route                                        | What fills the outlet |
| -------------------------------------------- | --------------------- |
| `/o/:o/p/:p/content`                         | first group's first view |
| `/o/:o/p/:p/content/c/:collection`           | the collection view   |
| `/o/:o/p/:p/content/*`                       | the editor            |

The editor keeps its current address, so links already sent keep working. The
collection view's search, folder, sort and page live in the query string, so
back and reload give the same list, and previous and next can be worked out
from it.

### Previous and next ask the server

Opening an item from a list carries the list's query in the address. The
editor asks the items route for the item's neighbours under that query
(`?around=<key>`), getting its index, the total and the keys either side. It
does not hold the whole list in memory, and a list that changed since is
answered as it is now.

### The server filters, sorts and pages

The items route takes `q`, `folder`, `sort` (`newest`, `oldest`, `title`,
`edited`), `page`, `pageSize` (20 by default) and `around`. It keeps
answering the whole list when none is given, as `content-follows-the-model`
left it. "Recently edited" asks history for each item's last version, once per
listing.

### Status and draft

`status` is changed, new or published as before, and `draft` when the item's
front matter has `draft: true`. The "Show on website" switch is the form's own
boolean control for the `draft` field, drawn inverted at the top and left out
of the field list below, so saving goes through the same code and the golden
tests hold.

### Making an item

`POST …/collections/{collection}/items` takes `title`, `folder` and `slug`
and answers 201 with the key, or a refusal with key
`refused.item-address-taken`. The dialog asks `GET …/collections/{collection}/address?title=&folder=`
while typing, which answers the slug, the web address, whether it is free,
and the file for those who may see technical detail. The address preview
uses the collection's `previewUrlBase` when the model has one, else the
folder.

### Quick find

`GET …/search?q=` answers up to 20 titles across singles and items the person
may read, grouped, dropping developer sections for those who may not see them.
⌘K or Ctrl+K opens it from any project screen; it is a dialog, so it works the
same in the phone frame.

### Remembered state

Group folding, the column fold and the file names switch are kept in local
storage under keys naming the person and project, wrapped so that a browser
refusing storage just forgets.

## Risks / Trade-offs

- [Reading front matter for every item on every query] → the listing is cached
  per site version, which changes only when something is saved or arrives,
  so a search keystroke reads no files.
- [Three routes where there was one] → the editor's address does not change,
  and the old content overview's address now opens the content layout.
- [Keyboard shortcuts clash with the browser or a field] → ⌘K or Ctrl+K and
  Alt+Up or Alt+Down are not taken by the browsers we support, and are ignored
  while typing in the Markdown editor.
