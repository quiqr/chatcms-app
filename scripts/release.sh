#!/usr/bin/env bash
# Cuts a release of ChatCMS. See docs/RELEASING.md.
#
#   scripts/release.sh patch|minor|major            release the next version
#   scripts/release.sh patch|minor|major --dry-run  say what it would do, change nothing
#
# It refuses, before changing anything, when the working copy has changes,
# when main differs from the remote, when [Unreleased] in CHANGELOG.md is
# empty, when the tag already exists, or when gh is not signed in.
#
# Then it writes VERSION, dates the [Unreleased] section, runs the gate, and
# commits, tags, pushes and publishes a GitHub release with that section as
# its notes. A failing gate puts both files back.
#
set -euo pipefail

bump="${1:-}"
dry_run=""
if [ "${2:-}" = "--dry-run" ]; then
    dry_run=1
fi

case "$bump" in
    patch | minor | major) ;;
    *)
        echo "usage: scripts/release.sh patch|minor|major [--dry-run]" >&2
        exit 2
        ;;
esac

cd "$(dirname "$0")/.."

refuse() {
    echo "release: $*" >&2
    exit 1
}

current="$(tr -d '[:space:]' < VERSION)"
if ! [[ "$current" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    refuse "VERSION holds '$current', which is not a bare semver"
fi
major="${BASH_REMATCH[1]}"
minor="${BASH_REMATCH[2]}"
patch="${BASH_REMATCH[3]}"

case "$bump" in
    patch) next="$major.$minor.$((patch + 1))" ;;
    minor) next="$major.$((minor + 1)).0" ;;
    major) next="$((major + 1)).0.0" ;;
esac
tag="v$next"
today="$(date +%F)"

section="$(awk '
    /^## \[Unreleased\]/ { inside = 1; next }
    inside && /^## \[/ { exit }
    inside {
        lines[++n] = $0
        if ($0 ~ /[^[:space:]]/) { if (!first) first = n; last = n }
    }
    END { for (i = first; first && i <= last; i++) print lines[i] }
' CHANGELOG.md)"

if ! grep -qvE '^(### .*)?[[:space:]]*$' <<<"$section"; then
    refuse "[Unreleased] in CHANGELOG.md is empty, so there is nothing to release"
fi

if [ -n "$(git status --porcelain)" ]; then
    refuse "the working copy has changes; ship or put them away first"
fi

git fetch --quiet --tags origin main
if [ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]; then
    refuse "the working copy is not on origin/main"
fi

if git rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
    refuse "tag $tag already exists"
fi

if ! gh auth status >/dev/null 2>&1; then
    refuse "gh is not signed in; run: gh auth login"
fi

echo "==> $current -> $next ($tag, $today)"
echo
echo "$section"
echo

if [ -n "$dry_run" ]; then
    echo "==> dry run: nothing changed"
    exit 0
fi

saved="$(mktemp -d)"
trap 'rm -rf "$saved"' EXIT
cp VERSION CHANGELOG.md "$saved/"

restore() {
    cp "$saved/VERSION" "$saved/CHANGELOG.md" .
}

echo "$next" > VERSION
awk -v header="## [$next] - $today" '
    { print }
    /^## \[Unreleased\]/ && !done { print ""; print header; done = 1 }
' "$saved/CHANGELOG.md" > CHANGELOG.md

echo "==> gate: nix flake check"
if ! nix flake check; then
    restore
    refuse "the gate failed; VERSION and CHANGELOG.md are put back"
fi

echo "==> commit"
jj commit -m "Release $next" VERSION CHANGELOG.md
jj bookmark set main -r @-
commit="$(jj log -r @- --no-graph -T commit_id)"

echo "==> tag $tag"
git tag -a "$tag" -m "ChatCMS $next" "$commit"

echo "==> push"
jj git push --bookmark main
git push origin "$tag"

echo "==> GitHub release"
printf '%s\n' "$section" > "$saved/notes.md"
if ! gh release create "$tag" --title "ChatCMS $next" --notes-file "$saved/notes.md"; then
    cp "$saved/notes.md" "./release-notes-$next.md"
    refuse "the tag is pushed, but the GitHub release failed; rerun:
    gh release create $tag --title \"ChatCMS $next\" --notes-file release-notes-$next.md"
fi

echo "==> released $tag"
