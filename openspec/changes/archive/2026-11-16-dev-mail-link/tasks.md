# Tasks

## 1. The inbox port reaches the frontend

- [x] 1.1 Export `CHATCMS_INBOX_PORT` from `scripts/dev-fe.sh` and list it in its help header. Verify with `scripts/dev-scripts.test.sh`: `CHATCMS_INSTANCE=test` hands the stub `npx` inbox port 8026, and the usual instance hands it 8025.
- [x] 1.2 Define `import.meta.env.VITE_CHATCMS_INBOX_PORT` in `web/vite.config.ts` from `CHATCMS_INBOX_PORT`, default `8025`, and declare its type. Verify with `npm run typecheck`.

## 2. The link

- [x] 2.1 Add `web/src/routes/DevMailLink.tsx` that links to `http://<hostname>:<port>` in a new tab, with strings in `en.ts` and `nl.ts`. Verify with a Vitest test that checks the href against `window.location.hostname` and the port.
- [x] 2.2 Show it in `DevSignIn` and in a lazily loaded development block on `ForgotPasswordPage`. Verify with Vitest: both screens show the link in development, and neither does with `DEV` stubbed false.
- [x] 2.3 Verify `npm run check:strings`, `npm run check:interface` and a production build followed by `npm run check:bundle` pass, and that `dist` contains no `DevMailLink` chunk.

## 3. Changelog

- [x] 3.1 Add the link to `CHANGELOG.md` under `[Unreleased]` → `Added`. Verify by reading the entry.

## Workflow follow-up

- Ship with `scripts/ship-change.sh dev-mail-link "<subject>"`.
- Add `openspec-link` to bean chatcms-rorq and close it with a Summary of Changes.
