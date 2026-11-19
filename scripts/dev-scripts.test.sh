#!/usr/bin/env bash
# shellcheck disable=SC2016
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

setup() {
    rm -rf "${work:?}/repo" "${work:?}/bin" "${work:?}/log"
    mkdir -p "$work/repo/scripts" "$work/repo/web/node_modules" "$work/bin" "$work/log"
    cp "$here/dev-instance.sh" "$here/dev-fe.sh" "$here/dev-be.sh" "$work/repo/scripts/"
    stub npx 'echo "args=$*" >"'"$work"'/log/npx"; echo "port=$CHATCMS_PORT host=$CHATCMS_HOST inbox=$CHATCMS_INBOX_PORT pwd=$PWD" >>"'"$work"'/log/npx"'
    stub npm 'echo "args=$* pwd=$PWD" >"'"$work"'/log/npm"'
    stub curl 'echo "$*" >"'"$work"'/log/curl"; exit "${CURL_EXIT:-0}"'
}

run_fe() {
    env -i PATH="$work/bin:$PATH" IN_NIX_SHELL=impure "$@" \
        "$bash_bin" "$work/repo/scripts/dev-fe.sh" 2>&1
}

instance_values() {
    env -i "$@" "$bash_bin" -c 'set -eu; root=/r; source "$0"; echo "$port $web_port $smtp_port $inbox_port $state $database $host"' \
        "$here/dev-instance.sh"
}

expect_eq "usual instance" "$(instance_values)" "8080 5173 1025 8025 /r/.chatcms-state chatcms_dev 127.0.0.1"
expect_eq "named instance" "$(instance_values CHATCMS_INSTANCE=test)" "8081 5174 1026 8026 /r/.chatcms-state-test chatcms_dev_test 127.0.0.1"
expect_eq "overridden ports" "$(instance_values CHATCMS_PORT=9000 CHATCMS_WEB_PORT=6000)" "9000 6000 1025 8025 /r/.chatcms-state chatcms_dev 127.0.0.1"
expect_eq "host" "$(instance_values CHATCMS_HOST=0.0.0.0 | awk '{print $7}')" "0.0.0.0"
expect_eq "AI gateway ports" "$(env -i "$bash_bin" -c 'set -eu; root=/r; source "$0"; echo "$ai_port"' "$here/dev-instance.sh") $(env -i CHATCMS_INSTANCE=test "$bash_bin" -c 'set -eu; root=/r; source "$0"; echo "$ai_port"' "$here/dev-instance.sh")" "8091 8092"

setup
out="$(run_fe)"
expect_contains "usual: vite port" "$(cat "$work/log/npx")" "args=vite --port 5173 --strictPort"
expect_contains "usual: api port" "$(cat "$work/log/npx")" "port=8080 host=127.0.0.1 inbox=8025 pwd=$work/repo/web"
expect_contains "usual: probe" "$(cat "$work/log/curl")" "http://127.0.0.1:8080/"
expect_absent "usual: no warning with a backend" "$out" "no backend"
[ -e "$work/log/npm" ] && fail "npm ci ran although node_modules exists"

setup
out="$(run_fe CHATCMS_INSTANCE=test CURL_EXIT=7)"
expect_contains "named: vite port" "$(cat "$work/log/npx")" "args=vite --port 5174 --strictPort"
expect_contains "named: api port" "$(cat "$work/log/npx")" "port=8081"
expect_contains "named: inbox port" "$(cat "$work/log/npx")" "inbox=8026"
expect_contains "named: warning" "$out" "no backend answers on port 8081"
expect_contains "named: warning names the backend" "$out" "CHATCMS_INSTANCE=test scripts/dev-be.sh"

setup
out="$(run_fe CHATCMS_HOST=0.0.0.0 CURL_EXIT=7)"
expect_contains "host: passed on" "$(cat "$work/log/npx")" "host=0.0.0.0"
expect_contains "host: probe on loopback" "$(cat "$work/log/curl")" "http://127.0.0.1:8080/"
expect_contains "host: warning keeps the host" "$out" "CHATCMS_HOST=0.0.0.0 scripts/dev-be.sh"

setup
env -i PATH="$work/bin:$PATH" IN_NIX_SHELL=impure \
    "$bash_bin" "$work/repo/scripts/dev-fe.sh" -- --open >/dev/null 2>&1
expect_contains "extra args" "$(cat "$work/log/npx")" "args=vite --port 5173 --strictPort --open"

setup
rm -rf "$work/repo/web/node_modules"
out="$(run_fe)"
expect_contains "fresh checkout: npm ci" "$(cat "$work/log/npm" 2>/dev/null)" "args=ci pwd=$work/repo/web"
[ -e "$work/log/npx" ] || fail "fresh checkout: vite did not start"

setup
status=0
out="$(env -i PATH="$work/bin:$PATH" "$bash_bin" "$work/repo/scripts/dev-fe.sh" 2>&1)" || status=$?
[ "$status" -ne 0 ] || fail "outside the dev shell: exited 0"
expect_contains "outside the dev shell" "$out" "nix develop"
[ -e "$work/log/npx" ] && fail "outside the dev shell: vite started"

bad_message="dev: an instance is named with small letters and digits"
setup
status=0
out="$(run_fe CHATCMS_INSTANCE=Bad!)" || status=$?
[ "$status" -ne 0 ] || fail "bad instance (fe): exited 0"
expect_contains "bad instance (fe)" "$out" "$bad_message"

status=0
out="$(env -i PATH="$work/bin:$PATH" CHATCMS_INSTANCE=Bad! "$bash_bin" "$work/repo/scripts/dev-be.sh" 2>&1)" || status=$?
[ "$status" -ne 0 ] || fail "bad instance (be): exited 0"
expect_contains "bad instance (be)" "$out" "$bad_message"

if [ "$failures" -ne 0 ]; then
    echo "$failures failure(s)" >&2
    exit 1
fi
echo "dev scripts: all tests passed"
