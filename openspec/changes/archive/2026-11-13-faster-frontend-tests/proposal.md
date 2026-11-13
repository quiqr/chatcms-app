# Proposal

## Why

The frontend check failed the gate on tests nobody had changed: ten screen
tests ran past Vitest's 5 second limit. On a quiet machine the slowest take 2
to 3 seconds; inside `nix flake check` they compete with the Go tests, the
integration tests, the VM test and Playwright, and Vitest starts a worker per
core on top of that. The previous change raised the limit to 30 seconds in the
check. This one makes the tests themselves cheaper, so the limit is a safety
net rather than the reason the gate passes.

Originating bean: none. It came up while shipping `header-everywhere`
(`chatcms-3op9`).

Acceptance test served: none. It is tooling for the gate.

## What Changes

- The `frontend` check runs Vitest with at most four workers, so it competes
  less with the other checks building at the same time. The 30 second limit
  stays, for a machine that is busy with other work.
- Tests that fill a field without being about typing do so in one step: a
  click and a paste, through a small `fill` helper. Typing a 33 character
  address one key at a time re-rendered the whole screen 33 times.
- Tests where typing is the point keep typing: the editor's one save rather
  than one per letter, the single character saves, the toolbar selection, and
  searching as you type.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. No behaviour of the product changes; `skip_specs` is set.

## Impact

- `flake.nix`: the `frontend` check's Vitest invocation.
- `web/src/routes/__tests__/`: a new `fill.ts`, and ten test files use it.
- `CHANGELOG.md`.

### Assumptions

- Four workers is a fair share of the machine during the gate; a local
  `npm run test` keeps Vitest's own default.
- A paste fires one input event where typing fires one per character. Every
  converted test reads the value after the field is filled, not along the
  way, so it proves the same thing.
