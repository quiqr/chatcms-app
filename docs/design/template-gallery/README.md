# Quiqr ChatCMS — template gallery hand-off

"Start from a template": a gallery of Quiqr community templates (from the
public `templates.json` feed) and the step that turns a template into a new
project.

## How to use it

1. Copy this folder into the repo, e.g. `docs/design/template-gallery/`.
2. Start Claude Code and paste the prompt from `PROMPT.md`.
3. `TEMPLATE-GALLERY.md` is the spec; `reference/` is the visual source of truth.

## What's inside

| Path | What it is |
| --- | --- |
| `TEMPLATE-GALLERY.md` | Spec: feed fields and URLs, backend caching + API, gallery, step 2, phone, roles, open decisions, acceptance checks. |
| `PROMPT.md` | Kick-off prompt with build order. |
| `tokens/tokens.css` | Design tokens (shadcn names, Tailwind v4). |
| `assets/quiqr-logo.svg` | Quiqr wordmark. |
| `reference/html/` | `gallery` (search works), `gallery-loading`, `gallery-error`, `use-template`, `phone-gallery`. |
| `reference/screenshots/` | Desktop 1440 px and phone 390 px PNGs. |
| `reference/canvas-source/project/` | The Claude Design artboards (`.dc.html`). |

## Notes

- Screenshots in the designs are coloured placeholders; real ones come from
  the feed. Only *Summer* uses real metadata; other descriptions are placeholders.
- Two decisions are assumed (spec §6): user-supplied empty repository, and a
  fresh git history for the copy.
