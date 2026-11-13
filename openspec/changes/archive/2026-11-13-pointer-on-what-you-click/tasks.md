# Tasks

## 1. The rule

- [x] 1.1 Add the pointer rule to `@layer base` in `web/src/index.css`, excluding disabled controls; test: `test/e2e/27-the-pointer.spec.mjs` finds the hand on "New project", "Help", the person menu and "Your account"
- [x] 1.2 Remove `cursor-default` from `MenuItem`; test: the same e2e test, on "Your account"
- [x] 1.3 A disabled button keeps the arrow; test: the disabled "Make it" button in the New project dialog in the same e2e test

## 2. The check

- [x] 2.1 `check-interface.mjs` fails on `cursor-default` or `cursor-auto` in any component, implementer screens included; test: a probe file setting `cursor-default` makes `npm run check:interface` fail and name it

## 3. Documents

- [x] 3.1 `docs/design/DESIGN.md` section 4 states the rule
- [x] 3.2 `CHANGELOG.md` under [Unreleased]
