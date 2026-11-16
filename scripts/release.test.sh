#!/usr/bin/env bash
# shellcheck disable=SC2016
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
bash_bin="$(command -v bash)"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
today="$(date +%F)"

export HOME="$work/home"
export GIT_AUTHOR_NAME=test GIT_AUTHOR_EMAIL=test@example.org
export GIT_COMMITTER_NAME=test GIT_COMMITTER_EMAIL=test@example.org
export GIT_CONFIG_NOSYSTEM=1
mkdir -p "$HOME"

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

expect_contains() {
    if ! grep -qF -- "$3" <<<"$2"; then
        fail "$1: expected to contain '$3', got: $2"
    fi
}

expect_absent() {
    if grep -qF -- "$3" <<<"$2"; then
        fail "$1: expected not to contain '$3', got: $2"
    fi
}

stub() {
    cat >"$work/bin/$1" <<EOF
#!$bash_bin
$2
EOF
    chmod +x "$work/bin/$1"
}

changelog='# Changelog

Some words.

## [Unreleased]

### Added

- A thing people can use.

### Fixed

- A thing that was broken.

## [0.1.4] - 2026-01-01

### Added

- Something older.'

setup() {
    local version="$1" log="${2:-$changelog}"
    rm -rf "${work:?}/origin.git" "${work:?}/repo" "${work:?}/bin" "${work:?}/calls"
    mkdir -p "$work/bin" "$work/repo/scripts"
    touch "$work/calls"
    git init -q --bare -b main "$work/origin.git"
    cp "$here/release.sh" "$work/repo/scripts/"
    cd "$work/repo"
    git init -q -b main
    git remote add origin "$work/origin.git"
    echo "$version" > VERSION
    printf '%s\n' "$log" > CHANGELOG.md
    git add -A
    git commit -qm start
    git push -q origin main
    cd "$here"

    stub nix 'echo "nix $*" >>"'"$work"'/calls"; exit "${NIX_EXIT:-0}"'
    stub gh 'echo "gh $*" >>"'"$work"'/calls"
case "$1 ${2:-}" in
    "auth status") exit "${GH_AUTH:-0}" ;;
    "release create")
        while [ $# -gt 0 ]; do
            if [ "$1" = "--notes-file" ]; then cp "$2" "'"$work"'/notes"; fi
            shift
        done
        exit "${GH_RELEASE:-0}" ;;
esac'
    stub jj 'echo "jj $*" >>"'"$work"'/calls"
case "$1" in
    commit) message="$3"; shift 3; git add "$@"; git commit -qm "$message" ;;
    log) git rev-parse HEAD ;;
    git) git push -q origin HEAD:main ;;
esac'
}

release() {
    (cd "$work/repo" && env PATH="$work/bin:$PATH" "$@" "$bash_bin" scripts/release.sh "${ARGS[@]}" 2>&1)
}

calls() {
    grep -vE '^gh auth status' "$work/calls" | awk '{print $1, $2}' | paste -sd, -
}

for case in "patch 0.1.5" "minor 0.2.0" "major 1.0.0"; do
    read -r bump want <<<"$case"
    setup 0.1.4
    ARGS=("$bump" --dry-run)
    out="$(release)"
    expect_contains "dry run $bump" "$out" "0.1.4 -> $want (v$want, $today)"
    expect_contains "dry run $bump: section" "$out" "- A thing people can use."
    expect_contains "dry run $bump: says so" "$out" "nothing changed"
    expect_eq "dry run $bump: VERSION" "$(cat "$work/repo/VERSION")" "0.1.4"
    expect_eq "dry run $bump: no calls" "$(calls)" ""
    expect_eq "dry run $bump: clean" "$(git -C "$work/repo" status --porcelain)" ""
done

setup 0.4.2
ARGS=(major --dry-run)
expect_contains "major resets" "$(release)" "0.4.2 -> 1.0.0"

setup 0.1.4
ARGS=(minor)
out="$(release)"
expect_contains "release: done" "$out" "released v0.2.0"
expect_eq "release: VERSION" "$(cat "$work/repo/VERSION")" "0.2.0"
expect_eq "release: order" "$(calls)" "nix flake,jj commit,jj bookmark,jj log,jj git,gh release"
head="$(git -C "$work/repo" rev-parse HEAD)"
expect_eq "release: tag on the remote" "$(git -C "$work/origin.git" rev-parse 'v0.2.0^{commit}')" "$head"
expect_eq "release: main on the remote" "$(git -C "$work/origin.git" rev-parse main)" "$head"
expect_eq "release: only two files" "$(git -C "$work/repo" show --name-only --format= HEAD | sort | paste -sd, -)" "CHANGELOG.md,VERSION"
expect_eq "release: commit subject" "$(git -C "$work/repo" log -1 --format=%s)" "Release 0.2.0"
expect_eq "release: changelog" "$(sed -n '5,9p' "$work/repo/CHANGELOG.md")" "## [Unreleased]

## [0.2.0] - $today

### Added"
expect_contains "release: older section kept" "$(cat "$work/repo/CHANGELOG.md")" "## [0.1.4] - 2026-01-01"
notes="$(cat "$work/notes")"
expect_eq "release: notes start" "$(head -1 <<<"$notes")" "### Added"
expect_contains "release: notes" "$notes" "- A thing that was broken."
expect_absent "release: notes stop at the next version" "$notes" "Something older"
expect_absent "release: notes have no header" "$notes" "Unreleased"
expect_contains "release: gh title" "$(cat "$work/calls")" "gh release create v0.2.0 --title ChatCMS 0.2.0"

refused() {
    local name="$1" want="$2"
    local status=0 out
    out="$(release)" || status=$?
    [ "$status" -ne 0 ] || fail "$name: exited 0"
    expect_contains "$name" "$out" "$want"
    expect_eq "$name: VERSION untouched" "$(cat "$work/repo/VERSION")" "0.1.4"
    expect_eq "$name: CHANGELOG untouched" "$(cat "$work/repo/CHANGELOG.md")" "$changelog"
    expect_eq "$name: no gate, commit or push" "$(calls)" ""
}

ARGS=(minor)

setup 0.1.4
echo stray > "$work/repo/stray"
refused "dirty working copy" "the working copy has changes"
rm "$work/repo/stray"

setup 0.1.4
git -C "$work/repo" commit -q --allow-empty -m ahead
refused "not on origin/main" "not on origin/main"

setup 0.1.4
git -C "$work/repo" tag v0.2.0
refused "existing tag" "tag v0.2.0 already exists"

setup 0.1.4
status=0
out="$(release GH_AUTH=1)" || status=$?
[ "$status" -ne 0 ] || fail "gh not signed in: exited 0"
expect_contains "gh not signed in" "$out" "gh is not signed in"

empty='# Changelog

## [Unreleased]

### Added

## [0.1.4] - 2026-01-01

- Something older.'
setup 0.1.4 "$empty"
status=0
out="$(release)" || status=$?
[ "$status" -ne 0 ] || fail "empty unreleased: exited 0"
expect_contains "empty unreleased" "$out" "nothing to release"

setup 0.1.4
status=0
out="$(release NIX_EXIT=1)" || status=$?
[ "$status" -ne 0 ] || fail "gate fails: exited 0"
expect_contains "gate fails" "$out" "the gate failed"
expect_eq "gate fails: VERSION put back" "$(cat "$work/repo/VERSION")" "0.1.4"
expect_eq "gate fails: CHANGELOG put back" "$(cat "$work/repo/CHANGELOG.md")" "$changelog"
expect_eq "gate fails: stops at the gate" "$(calls)" "nix flake"
expect_eq "gate fails: no tag" "$(git -C "$work/repo" tag)" ""

setup 0.1.4
status=0
out="$(release GH_RELEASE=1)" || status=$?
[ "$status" -ne 0 ] || fail "gh release fails: exited 0"
expect_contains "gh release fails: rerun" "$out" "gh release create v0.2.0"
expect_contains "gh release fails: tag pushed" "$(git -C "$work/origin.git" tag)" "v0.2.0"
[ -f "$work/repo/release-notes-0.2.0.md" ] || fail "gh release fails: notes kept"

setup 0.1.4
ARGS=(sideways)
status=0
release >/dev/null || status=$?
expect_eq "bad usage" "$status" "2"

if [ "$failures" -ne 0 ]; then
    echo "$failures failure(s)" >&2
    exit 1
fi
echo "release: all tests passed"
