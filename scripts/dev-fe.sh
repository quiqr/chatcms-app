#!/usr/bin/env bash
# Starts the frontend development server, with live reloading, for the
# backend that scripts/dev-be.sh starts. Run it inside `nix develop`.
#
#   scripts/dev-fe.sh                 serve on :5173, API requests go to :8080
#   scripts/dev-fe.sh -- <vite args>  anything after the script name goes to Vite
#
# It installs the frontend dependencies when web/node_modules is missing, and
# warns, but starts anyway, when no backend answers on the API port.
#
# The sign-in screen has a button for each seeded role: administrator,
# implementer, editor and viewer. A built binary does not.
#
# It reads the same settings as scripts/dev-be.sh, so the same words in front
# of both start two halves that belong together:
#
#   CHATCMS_INSTANCE=test scripts/dev-fe.sh   serve on :5174, API on :8081
#   CHATCMS_HOST=0.0.0.0 scripts/dev-fe.sh    reachable from a phone
#
#   CHATCMS_INSTANCE    empty for the usual instance, or small letters and digits
#   CHATCMS_HOST        address to listen on, default 127.0.0.1
#   CHATCMS_PORT        API port, default 8080, or 8081 for a named instance
#   CHATCMS_WEB_PORT    frontend port, default 5173, or 5174 for a named instance
#
set -euo pipefail

cd "$(dirname "$0")/.."
root="$PWD"

# shellcheck source=scripts/dev-instance.sh
source "$root/scripts/dev-instance.sh"

if [ -z "${IN_NIX_SHELL:-}" ] || ! command -v npm >/dev/null; then
    echo "dev-fe: run this inside 'nix develop', which pins Node." >&2
    exit 1
fi

if [ ! -d "$root/web/node_modules" ]; then
    echo "==> installing the frontend dependencies"
    (cd "$root/web" && npm ci)
fi

probe="$host"
if [ "$probe" = "0.0.0.0" ] || [ "$probe" = "::" ]; then
    probe="127.0.0.1"
fi

if ! curl -s -o /dev/null --max-time 1 "http://$probe:$port/"; then
    echo "dev-fe: no backend answers on port $port; start it in another terminal with:" >&2
    echo "    ${instance_prefix}scripts/dev-be.sh" >&2
fi

if [ "${1:-}" = "--" ]; then
    shift
fi

export CHATCMS_PORT="$port"
export CHATCMS_HOST="$host"

cd "$root/web"
exec npx vite --port "$web_port" --strictPort "$@"
