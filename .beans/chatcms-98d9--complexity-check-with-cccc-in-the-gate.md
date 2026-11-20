---
# chatcms-98d9
title: Complexity check with cccc in the gate
status: todo
type: task
priority: normal
created_at: 2026-11-20T07:18:47Z
updated_at: 2026-11-20T07:18:47Z
---

Measure the cognitive and cyclomatic complexity of every function with
[cccc](https://github.com/moznion/cccc), in Go (`--lang go`) and in the
frontend's TypeScript/TSX (`--lang es`).

- Package cccc in the flake (a Rust build of `cccc-cli`, pinned).
- A first run over `internal/`, `cmd/` and `web/src`: which functions are
  the most complex (`--top-cognitive`), and what limit is realistic.
- A `cccc.toml` with `max-cognitive` (and possibly `max-cyclomatic`),
  excluding generated code (`*.generated.go`) and tests if that makes sense.
- A check in `flake.nix` and in the right tier of `scripts/gate.sh`, so a
  function above the limit fails the gate.
- Functions above the limit today are either simplified or listed as known
  exceptions, so the check starts green.
