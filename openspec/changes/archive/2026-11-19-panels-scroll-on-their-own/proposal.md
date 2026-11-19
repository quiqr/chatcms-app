# Proposal

Bean: [chatcms-x9uq](../../../.beans/chatcms-x9uq--form-should-scroll-not-the-whole-chatcms-interace.md)
(form should scroll, not the whole interface), milestone 03.1
([chatcms-iqog](../../../.beans/chatcms-iqog--ui-cleanups.md)). Closes
[chatcms-y9bd](../../../.beans/chatcms-y9bd--keep-header-fixed-when-scrolling-down.md)
(keep header fixed when scrolling down) as well.

## Why

Every frame is at least as tall as the window and the page itself scrolls.
In the editor the form, the preview and the assistant stand in one row as
tall as the form, so with a long form the bottom of the preview and the
assistant's message box end up below the fold, and the header scrolls away.
It serves acceptance test 5 of `docs/BRIEFING-M01.md`: a change only shows
"within a few seconds" if the preview is in view while the form is edited.

## What Changes

- **The page no longer scrolls; each panel scrolls on its own.** Every
  desktop frame is exactly as tall as the window. The header never scrolls.
  The navigation, the content menu, the form, the preview and the assistant
  each scroll inside themselves, or fill the height when they do not need
  to.
- **The editor shows what fits, and the view switch chooses the rest.** Where
  form and preview fit side by side, both show. Where they do not, they no
  longer stack: the existing Form / Side by side / Preview switch, now shown
  at every desktop width, chooses one, and the form comes first.
- **The assistant takes the form's place when there is no room for a third
  column**, so the preview with the assistant's result stays beside it.
- **A narrow window folds the panels rather than stacking them.** Too narrow
  for navigation beside content, the navigation shows as its rail and the
  content menu as its strip, rather than moving above the content.
- The phone frame does not change: it already gives its preview and chat a
  fixed height and keeps its bar at the bottom.

## Capabilities

### New Capabilities

- `interface/scrolling`: what scrolls in a frame and what stays put.

### Modified Capabilities

- `interface/app-shell`: "The frame reflows rather than breaking": a narrow
  window folds the navigation and the content menu instead of moving them
  above the content.

## Impact

- `web/src/shells/`: `EditorShell`, `ProjectShell`, `OrganisationShell`,
  `OrganisationAdminShell`, `PlatformAdminShell`, `SitePanel` and the content
  panel: the frame's height and which element scrolls.
- `web/src/routes/EditorPage.tsx`, `Preview.tsx`, `ChatPanel.tsx`: fill the
  height they are given instead of the form's.
- The view switch's visibility; the e2e and screen checks at 1440 and 390
  px, plus a medium width where the editor would have stacked.
- Touches none of gitsync, runner, policy or auth.

## Assumptions

- **One behaviour at every desktop width**, not a frame when wide and page
  scrolling when medium: what scrolls should not depend on the window size.
- **The breakpoint for "fits side by side"** is where form and preview each
  keep their current minimum of 420 px beside the open panels; below it the
  switch chooses. The choice lasts while the editor is open, as the layout
  does now; remembering it is not part of this change.
- **The assistant replaces the form, not the preview, when it does not fit
  as a column**, and the form comes back when it closes; on a wide window it
  stays a column of its own. (First written the other way round; the
  assistant's own acceptance tests showed the preview is what it needs.)
- **Folding when narrow does not overwrite the person's own fold state**; a
  panel they unfold on a narrow window lies over the content and folds back
  when they pick something.
- **The phone frame is left as it is.**
- **chatcms-y9bd is done by this change**, since a header outside the
  scrolling area cannot scroll away.
- **Fixed on the way: chatcms-koz4.** The e2e harness picked a free port and
  started the server later, so two tests could pick the same one and one
  server failed to listen. The harness now starts its server again on fresh
  ports when the port turns out to be taken; it cost three gate runs today.
