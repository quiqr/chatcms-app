#!/usr/bin/env bash
# Every changelog entry is at most two lines of at most 80 characters, so it
# says what was done and leaves the reasoning to the OpenSpec archive.
#
#   scripts/check-changelog.sh [CHANGELOG.md]
set -euo pipefail

file="${1:-$(dirname "$0")/../CHANGELOG.md}"

awk -v file="$file" '
    function close_entry() {
        if (start && lines > 2) {
            printf "%s:%d: entry is %d lines, at most 2\n", file, start, lines
            bad = 1
        }
        start = 0
    }
    /^- / { close_entry(); start = NR; lines = 1 }
    /^  [^ ]/ && start { lines++ }
    !/^- / && !/^  [^ ]/ { close_entry() }
    (start) && length($0) > 80 {
        printf "%s:%d: line is %d characters, at most 80\n", file, NR, length($0)
        bad = 1
    }
    END { close_entry(); exit bad }
' "$file"

echo "check-changelog: every entry is two lines or fewer"
