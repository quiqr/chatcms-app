#!/usr/bin/env bash
set -euo pipefail

nix_bin="${NIX:-nix}"
fast=(go-build go-test frontend sqlc changelog release dev-scripts)
ship=("${fast[@]}" integration e2e)

usage() {
    echo "usage: gate.sh <fast|ship|release>" >&2
    echo "  fast     ${fast[*]}" >&2
    echo "  ship     ${ship[*]}" >&2
    echo "  release  nix flake check, every check" >&2
    exit 2
}

build() {
    local system
    system="$("$nix_bin" eval --impure --raw --expr builtins.currentSystem)"
    local targets=()
    for check in "$@"; do
        targets+=(".#checks.${system}.${check}")
    done
    echo "==> gate: ${*}"
    if ! "$nix_bin" build --no-link -L "${targets[@]}"; then
        echo "gate: a check failed; the build log above names it" >&2
        exit 1
    fi
}

case "${1:-}" in
    fast) build "${fast[@]}" ;;
    ship) build "${ship[@]}" ;;
    release)
        echo "==> gate: nix flake check"
        "$nix_bin" flake check
        ;;
    *) usage ;;
esac
