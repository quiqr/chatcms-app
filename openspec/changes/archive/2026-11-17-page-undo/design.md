# Design

## Context

`PageEditor.tsx` holds the page as controls with values, collects changes in
`waiting` (values by path, and the body), and writes them with `savePage`
800 ms after the last change. A save that fails sets the trouble state and
drops what was waiting. Nothing records what a field held before a change, so
there is nothing to undo with. See proposal.md for why.

## Goals / Non-Goals

**Goals:**

- The rules of undo (record, undo, redo, go to a point, what to leave alone,
  what to keep) live in one pure module that is tested without a screen.
- An undo goes through the same save as an edit, so the server and the
  version history see nothing new.

**Non-Goals:**

- Undo across pages, or of renames, copies, removals and pictures uploaded in
  the media library. Those are not edits made in the form.
- Page history from earlier days, and hiding the editor from viewers.

## Decisions

### A pure module for the steps

`web/src/routes/undo.ts` holds the types and functions, with no React:

```ts
type Change = { before: unknown; after: unknown };
type Step = { at: string; values: Record<string, Change>; body?: Change };
type Steps = { done: Step[]; undone: Step[] };
```

- `record(steps, step)` adds to `done`, empties `undone`, keeps 100.
- `planUndo(steps, now)` and `planRedo(steps, now)` take the current values
  and return what to write, which fields to leave alone, and the steps after.
  A field is left alone when its current value is not deeply equal to the
  value the step left in it.
- `planTo(steps, index, now)` folds several undos or redos into one write, so
  that picking a row is one save.
- `fits(steps, page)` says whether a loaded page still matches where the
  steps left it, comparing only the paths the steps touch.

Alternative: keep the steps in React state inside `PageEditor`. Rejected,
because the rules are where the mistakes would be, and a pure module can be
tested on every case cheaply.

### Before values come from the screen

When a path first enters `waiting`, `PageEditor` notes its current value
from the controls (found by path, as `withValue` already walks them) as the
step's before; the body the same. When `savePage` succeeds, the waiting
changes become one step. When it fails, no step is made.

### Undo writes past the edit path

Undo, redo and picking a row use a second way to write: it first awaits any
save in flight and writes what is waiting (which records a step), then plans,
updates the controls and calls `savePage` with the planned values. It does not
pass through `change`, so it neither records a step of its own nor empties the
redo list. Saves are made one at a time, so an undo cannot overtake an edit.

### Where steps are kept

sessionStorage under `chatcms.undo.<person>.<project>.<page>`, through
try/catch like `lib/remember.ts`, holding `Steps`. On opening a page the
stored steps are used only when `fits` says the loaded page matches;
otherwise they are dropped. sessionStorage belongs to the tab, which is the
lifetime the spec asks for. Without storage the steps last until the page is
left.

### The controls

In the editor's header row, beside the saved status: Undo (`Undo2`), Redo
(`Redo2`) and Undo history (`ListRestart`), as icon buttons of 44 px. Undo's
accessible name says what it would undo ("Undo Title, Summary"). The history
is a `Menu`: one item per step, newest first, then "Opened"; steps that can be
redone are shown above the current point and dimmed; the current point is
checked. Times come from `lib/when.ts`. Field names come from the controls'
`title`; the page text uses the existing `page-editor.page-text` string.

A field left alone is said with the existing `toast`: "Summary was changed
since, so it was left as it is."

### Keys

A window `keydown` in `PageEditor`: Ctrl+Z or Cmd+Z undoes, with Shift or as
Ctrl+Y it redoes, unless the target is a text field. The `typing` check that
`usePlace.tsx` already has for Alt+Up and Alt+Down moves to `lib/typing.ts`
and is used by both.

## Risks / Trade-offs

- [Deep comparison of large values, such as a long list field, on every undo]
  → it runs on one press over the fields of one step; a few kilobytes of JSON
  at most.
- [A save in flight when Undo is pressed] → undo awaits it first, so the
  order on the server is the order on the screen.
- [Rich fields whose value the server normalises, so "after" never matches
  what comes back] → the step stores what was sent and the screen keeps what
  was sent; the editor does not re-read after a save, so the two agree.

## Migration Plan

None. No stored data exists before this change, and removing it leaves only
unused sessionStorage entries that the tab drops.
