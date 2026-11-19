# Design

## Context

See proposal.md. Every desktop shell is a `flex min-h-dvh flex-wrap` row, so
the document scrolls. In the editor (`EditorShell`) the form, preview and
assistant slots sit in `main.flex-wrap` and stretch to the tallest of them.
The content menu already folds to its strip below `wideWidth` (1024 px,
`useIsWide`) and opens as a drawer there, without touching the person's fold
(`useFold`). The navigation (`SitePanel`) does not. The view switch exists in
`EditorShell` but is hidden below `md`, and the layout is plain state in
`EditorPage`.

## Goals / Non-Goals

**Goals:** one scrolling model for every desktop shell; the editor never
stacks; the assistant's box always in view.

**Non-Goals:** the phone frame; remembering the editor layout; the folded
menu's label strip and location picker (chatcms-vqsf); a fixed column width
system.

## Decisions

### 1. The frame: a window-tall grid with scrolling cells

Each desktop shell becomes `h-dvh overflow-hidden` with no wrap. The column
holding the top bar is `flex flex-col min-h-0`; the top bar is `shrink-0`;
`main` is `flex-1 min-h-0`. Panels that hold lists or forms get
`min-h-0 overflow-y-auto`. The preview and the assistant get `h-full` of a
`min-h-0` parent, so they take the window's height, not the form's.

Alternative: sticky header, preview and assistant over a scrolling page.
Rejected: it keeps two models (stacking still needs the page to scroll) and
long panels would need their own scroll inside the sticky anyway.

### 2. Fitting the editor's slots

A hook `useEditorRoom()` measures the slot row (`ResizeObserver` on `main`)
and answers what fits: `both` when the row is at least 2 × 420 px, plus
340 px when the assistant is open as a column; otherwise `one`. With `one`,
the shown layout is the person's choice if it is `form` or `preview`, and
`form` if they chose `split`. The switch shows at every desktop width;
`split` is offered only where it fits.

Alternative: media queries on the window. Rejected: the room depends on the
two side panels and their folds, not on the window.

### 3. The assistant on a medium window

When the assistant opens and there is no room for a 340 px column beside what
is shown, it takes the form's place and the preview stays: what the assistant
changes shows in the preview, and select-to-edit starts there. Where not even
preview and assistant fit (under 760 px), the assistant shows alone. It is the
same `ChatPanel`, kept mounted as now, so the conversation survives.

Alternative: the assistant in the preview's place, keeping the form. Tried
first, and rejected after the assistant's acceptance tests: on a 1440 px
window with both side panels open there is no room for three columns, and
hiding the preview there hid the assistant's own result.

### 4. Narrow windows fold the navigation like the content menu

`SitePanel` follows `ContentPanel`: below `wideWidth` it shows as its rail;
unfolding opens it as an overlay drawer that closes on a pick; the person's
`useFold` state is read but not written by the narrow state.

### 5. Other shells

`ProjectShell`, `OrganisationShell`, `OrganisationAdminShell` and
`PlatformAdminShell` get the same frame: the top bar stays, the page's main
area scrolls. Screens that set their own `min-h-dvh` drop it.

## Risks / Trade-offs

- **Nested scroll areas** can trap the wheel inside a short panel: every
  scrolling cell gets `overscroll-contain` only where it matters (the form
  and the assistant's list), so the page never needs to scroll anyway.
- **Screens that relied on page scrolling** (anchors, `scrollIntoView` on the
  window) move to scrolling their panel; the e2e checks at both widths catch
  a screen that clips.
- **Playwright's `toBeVisible`** passes for an element scrolled out of its
  panel; the new checks assert positions inside the viewport instead.
