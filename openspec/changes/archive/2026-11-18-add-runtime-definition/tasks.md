# Tasks

## 1. The runtime definition

- [x] 1.1 `internal/runtime`: parse, validate, `Default`, `Fill`; proven by
  unit tests (format, unknown keys, types, placeholders, a shell trick as one
  word, no preview command).
- [x] 1.2 Resolving packages and programs through `nixpin`, and without Nix;
  the app's `runtimes.For` with the fallback to the recorded Hugo; proven by
  unit tests with fake pins.

## 2. Previews run what is declared

- [x] 2.1 `Settings.RuntimeFor`, `Spec.Run`; the local backend fills and
  runs it; proven by manager and local backend tests.
- [x] 2.2 The microvm backend passes `chatcms.run`; the guest runs it;
  proven by the backend tests and the hand-run boot test.

## 3. Builds run what is declared

- [x] 3.1 `sitebuild` runs the resolved build and its output; `Pages` reads
  `Root`; proven by sitebuild tests.

## 4. The preview origin

- [x] 4.1 View keys in the manager; the preview origin's routing and checks;
  the app origin's `/preview/` closed when an origin is set; proven by
  preview and httpapi tests.
- [x] 4.2 `CHATCMS_PREVIEW_URL` and the NixOS module's `previewUrl`; proven
  in `config_test.go`.

## 5. The pane

- [x] 5.1 `address`, `type`, `open` from the workspace; viewers for web, pdf
  and other; proven by Vitest.
- [x] 5.2 e2e: a site with a runtime definition previews from a separate
  origin; a pdf preview shows in the pane; at both widths.

## 6. Ship

- [x] 6.1 `docs/ARCHITECTURE.md`, `docs/DEPLOYING.md`, a runtime definition
  reference in `docs/`, `CHANGELOG.md`.
- [x] 6.2 `scripts/gate.sh release` and `scripts/gate.sh ship` green.
