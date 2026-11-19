---
# chatcms-gf1r
title: 'SSG-agnostic anchors: find a selection in the page source, Hugo as tier 1'
status: in-progress
type: feature
priority: normal
created_at: 2026-11-19T20:04:29Z
updated_at: 2026-11-20T01:00:00Z
openspec-link: openspec/changes/archive/2026-11-19-where-a-selection-is-from
parent: chatcms-z0p7
---

## Purpose

ChatCMS should not depend on one static site generator. The chat design
(`docs/design/chat-requests/CHAT.md` §2, §6) maps a selection in the preview
to a field through `data-quiqr-field` and `data-quiqr-tpl` attributes that a
preview build adds, which only works where we control the templates.

## Decision (19 November 2026, with Pim)

Two levels, one answer:

- **Generic, any SSG**: the page comes from the preview address (as today);
  the field comes from searching the selected text and its context in the
  page's source, the front matter values and the body, which the model
  already describes. One match gives the field; several are resolved by
  context or left to the assistant; none suggests the template or shared
  content.
- **Tier 1, Hugo**: preview builds add `data-quiqr-field` and
  `data-quiqr-tpl` (render hooks and a partial) so the answer is exact:
  field, shared data, template file and line.

An SSG is then a set of capabilities rather than supported or not; the
shortcode dialects follow the same pattern.

## Scope

- The generic matcher on the server, with tests over the fixtures.
- The Hugo markers in preview builds, as a spike first: whether
  `templates.Current` in our Hugo version gives file and line.
- The anchor the overlay sends carries what either level found.

Needed for CHAT.md §6 (content, shared or template).
