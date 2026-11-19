# Proposal

Bean: [chatcms-gf1r](../../../.beans/chatcms-gf1r--ssg-agnostic-anchors-find-a-selection-in-the-page.md)
(SSG-agnostic anchors: find a selection in the page source, Hugo as tier 1),
milestone 07 ([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).
Design: `docs/design/chat-requests/CHAT.md` §6.

## Why

A selection in the preview reaches the assistant with the page it came from,
and only on a Hugo site, through Hugo's own list of addresses. Which field
holds the text, whether it is shared text, or whether it comes from the
site's design is left for the assistant to guess. ChatCMS should not depend
on one generator for this, and §6 of the chat design needs the answer to
show the right card.

## What Changes

- **A generic answer, for any generator**: the selected text and its
  context are looked up in the source the model describes: the front matter
  values and the body of the site's pages. The result is a field of a page,
  shared text (a single whose file is outside `content/`, such as
  `data/settings.yaml`), or the site's design when no page holds the text.
- **Hugo first**: where Hugo's list of addresses names the page, that page is
  searched first; the generic answer covers everything else.
- **The assistant is told** which field of which page, that the text is
  shared and shows on more pages, or that it comes from the design and
  cannot be changed.
- **`POST …/chat/where`** answers the same question for a selection, for the
  cards of §6 in the next change.

## Capabilities

### New Capabilities

### Modified Capabilities

- `assistant/selection`: the assistant is told the field, shared text or the
  design, found in the source for any generator.

## Impact

- `internal/anchor`: the matcher, with its own tests.
- `internal/assistant`: the prompt; `Where`.
- `internal/httpapi`: `POST …/chat/where`.
- `internal/app`: the placer reads the person's copy of the site.

## Assumptions

- **Tier 1 markers** (`data-quiqr-field`, `data-quiqr-tpl` in Hugo preview
  builds, for exact fields and template lines) are a later step, after a
  spike; the bean stays open for them.
- **Matching** compares plain text: Markdown marks, links, HTML tags,
  entities and shortcodes are taken out, whitespace collapsed, case ignored.
  Where several fields hold the text, the one whose text also holds the
  selection's context wins.
- **Data files** (YAML, JSON, TOML) are not read as forms by the app yet;
  for this question they are read line by line as `key: value`.
- A site is searched up to 1000 pages.
