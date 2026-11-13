# Design

## Context

Measured on the development machine while idle, the slowest test took 3.3
seconds in a full run and 0.7 seconds on its own: most of the cost is workers
competing, not the test. Pasting brought the same test to 0.24 seconds alone.
The slowest tests that remain do not type: they are the first test of their
file and carry the cost of loading the screen and its jsdom environment.

## Goals / Non-Goals

**Goals:**
- The gate's frontend tests finish well inside the default limit on a machine
  doing nothing else.

**Non-Goals:**
- Changing what any test proves.
- Splitting test files or changing the test environment.

## Decisions

**Cap workers in the check, not in `vite.config.ts`.** A local run is the
only thing running and may use every core. In the gate the frontend check is
one of several builds at once, and a worker per core oversubscribes the
machine. The cap is a command-line flag in the check.

**One helper instead of an inline click and paste.** `fill(field, text)`
says what the test means. It uses the module-level `userEvent` API, so it
works the same in tests that use a `userEvent.setup()` instance and in tests
that do not.

**Keep the 30 second limit.** Other work on the same machine is outside the
gate's control; the limit stops that from failing it, and the faster tests
keep it from mattering otherwise.

## Risks / Trade-offs

- [A converted test would have caught a bug that only shows while typing] →
  The tests where typing is the behaviour under test keep typing.
