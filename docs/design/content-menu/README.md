# Quiqr ChatCMS — content menu hand-off (direction B)

The chosen design for the content menu: a content column built from the
Quiqr model menu, collection views (e.g. blog posts) and an editor that keeps
the menu next to the open page.

## How to use it

1. Copy this folder into the frontend repo, e.g. `docs/design/content-menu/`.
2. Start Claude Code in the repo and paste the prompt from `PROMPT.md`.
3. `CONTENT-MENU.md` is the spec Claude Code works from; the screenshots and
   HTML in `reference/` are the visual source of truth.

## What's inside

| Path | What it is |
| --- | --- |
| `CONTENT-MENU.md` | Spec: model mapping, layout, content column, collection view, new-item dialog, editor header, roles, API shape, acceptance checks. |
| `PROMPT.md` | Kick-off prompt with a step-by-step build order. |
| `tokens/tokens.css` | Same design tokens as the main kit (shadcn variable names, Tailwind v4). |
| `assets/quiqr-logo.svg` | Quiqr wordmark. |
| `reference/html/` | Static pages: `content` (collection view), `content-no-results`, `content-new-post`, `editor`, `editor-menu-hidden`, `phone-content-menu`, `phone-collection`. Linked: New post → dialog → Cancel; fold/unfold the menu in the editor. |
| `reference/screenshots/desktop/` | 1440 px full-page PNGs. |
| `reference/screenshots/phone/` | 390 px (2×) PNGs. |
| `reference/canvas-source/project/` | The four Claude Design artboards for direction B (`.dc.html`), to keep designing in a canvas. |

## Notes

- The desktop editor layout is for ≥ 1024 px; below that the content column
  becomes a drawer, and phones use the phone screens (see spec §2).
- Sample content (technative.eu-like groups, posts, people) is made up.
- Builds on the milestone-1 design kit (`DESIGN.md`) for everything outside
  the content area.
