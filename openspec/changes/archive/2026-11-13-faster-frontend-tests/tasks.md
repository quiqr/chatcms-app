# Tasks

## 1. The gate

- [x] 1.1 Run Vitest in the `frontend` check with `--maxWorkers=4`, keeping `--testTimeout=30000`; test: `nix flake check` passes

## 2. The tests

- [x] 2.1 Add `web/src/routes/__tests__/fill.ts`: click the field, paste the text
- [x] 2.2 Use `fill` in the tests that fill a field without being about typing (administration, content, editor's list item, media, onboarding, project settings, publish, sign-in, tenancy, whole site); test: all 213 Vitest tests pass and `npm run typecheck` is clean
- [x] 2.3 Leave typing in the editor's save tests, the toolbar test and the two search tests

## 3. Documents

- [x] 3.1 `CHANGELOG.md` under [Unreleased]
