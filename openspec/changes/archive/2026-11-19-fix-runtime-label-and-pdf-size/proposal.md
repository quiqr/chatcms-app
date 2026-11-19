# Proposal

No bean: two things found by hand with the seeded Quarto book, milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

- Every project card said "Hugo", the Quarto book's too: it was written
  into the card.
- A PDF preview was scaled to the pane's width, and narrowed further by the
  phone setting, while a PDF has a fixed size of its own.
- A Quarto preview on the local backend could miss its start: the backend
  waited 30 seconds for an answer, and Quarto renders the whole book, code
  cells included, before it answers.

## What Changes

- A project says what it runs on: the runtime definition's program, Hugo
  for a site without one, nothing for a project without a site.
- A PDF preview shows its pages at their own size, scrolling where the pane
  is smaller, with no choice of width.
- The local backend waits up to two minutes for a preview to answer, as the
  workspace manager already allows.

## Capabilities

### New Capabilities

### Modified Capabilities

- `preview/pane`: a PDF at its own size.
- `tenancy/projects`: what a project runs on.

## Impact

- `internal/tenancy`, `internal/app` (`siteRuntimes`), `internal/httpapi`.
- `internal/workspace/local`: `DefaultReady`.
- `web/src/routes/ProjectsPage.tsx`, `Preview.tsx`, `web/src/preview/pdf/`.

## Assumptions

- A PDF page is shown at 96/72: one PDF point is one CSS point.
