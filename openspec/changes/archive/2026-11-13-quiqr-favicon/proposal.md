# Proposal

## Why

The application names no icon, so the browser asks for `/favicon.ico`, gets
the application's page back, and shows its own placeholder in the tab. The
tab should show the Quiqr icon, the same one quiqr.org uses.

Originating bean: `chatcms-n742` (wrong favicon).

Acceptance test served: none.

## What Changes

- `web/public/favicon.png`: the icon quiqr.org serves at
  `/images/logo/02_QUIQR-icon-logo.png`, a 293 by 293 PNG, copied unchanged.
- `web/index.html` names it as the icon and as the icon for a phone's home
  screen.
- The bundle check fails when the built page does not name the icon or the
  icon is missing from the build.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `interface/design-system`: the browser tab carries the Quiqr icon.

## Impact

- `web/public/favicon.png`, `web/index.html`, `web/scripts/check-bundle.mjs`.
- `CHANGELOG.md`.
- No server change: the server already serves any file in the build.

### Assumptions

- One PNG is enough. Every current browser accepts a PNG icon named in the
  page, which is what quiqr.org does too; no `favicon.ico` is added.
- The icon is Quiqr's own artwork and may be used in Quiqr's application.
