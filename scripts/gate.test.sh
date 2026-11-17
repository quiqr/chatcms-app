#!/usr/bin/env bash
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
bash_bin="$(command -v bash)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

failures=0
fail() {
    echo "FAIL: $*" >&2
    failures=$((failures + 1))
}

expect_eq() {
    if [ "$2" != "$3" ]; then
        fail "$1: expected '$3', got '$2'"
    fi
}

stub_nix() {
    cat >"$work/nix" <<STUB
#!$bash_bin
if [ "\$1" = eval ]; then
    echo -n x86_64-linux
    exit 0
fi
echo "\$*" >>"$work/asked"
exit $1
STUB
    chmod +x "$work/nix"
    : >"$work/asked"
}

run() {
    NIX="$work/nix" bash "$here/gate.sh" "$@" >"$work/out" 2>&1
}

checks_of() {
    tr ' ' '\n' <"$work/asked" | sed -n 's/^\.#checks\.x86_64-linux\.//p' | tr '\n' ' ' | sed 's/ $//'
}

stub_nix 0
run fast || fail "fast: exited non-zero"
expect_eq "fast asks for its checks" "$(checks_of)" "go-build go-test frontend sqlc changelog release dev-scripts"

stub_nix 0
run ship || fail "ship: exited non-zero"
expect_eq "ship adds integration and e2e" "$(checks_of)" \
    "go-build go-test frontend sqlc changelog release dev-scripts integration e2e"

stub_nix 0
run release || fail "release: exited non-zero"
expect_eq "release runs flake check" "$(cat "$work/asked")" "flake check"

stub_nix 1
if run ship; then
    fail "a failing check: the gate passed"
fi
if ! grep -q "a check failed" "$work/out"; then
    fail "a failing check: not said"
fi

stub_nix 0
status=0
run quick || status=$?
expect_eq "an unknown tier exits 2" "$status" "2"
if ! grep -q "fast" "$work/out" || ! grep -q "release" "$work/out"; then
    fail "an unknown tier: the tiers are not named"
fi
expect_eq "an unknown tier runs nothing" "$(cat "$work/asked")" ""

if [ "$failures" -gt 0 ]; then
    echo "$failures failure(s)" >&2
    exit 1
fi
echo "gate.test.sh: all good"
