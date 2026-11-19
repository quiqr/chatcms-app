#!/usr/bin/env bash
# Starts everything needed to click through ChatCMS on this machine: a local
# PostgreSQL, a master key, a seeded database with one site that has a remote
# of its own on disk, Mailpit, and the server. Run it inside `nix develop`.
#
#   scripts/dev-be.sh          start the server, seeding only the first time
#   scripts/dev-be.sh fresh    throw this instance's database and state away
#   scripts/dev-be.sh stop     stop a server left running, and the database
#
# Ctrl-C stops the server. The database keeps running, because the next start
# is quicker with it up; "stop" puts it away as well.
#
# Then run the frontend in another terminal, with the same settings in front:
#
#   scripts/dev-fe.sh
#
# To reach it from another machine, such as a phone on the same network, give
# it an address to listen on. Both halves read the same setting:
#
#   CHATCMS_HOST=0.0.0.0 scripts/dev-be.sh
#   CHATCMS_HOST=0.0.0.0 scripts/dev-fe.sh
#
# Anything on the network can then reach it, with a password written in the
# seed, so do that on a network you trust and not on a shared one.
#
# A second instance, so that two people can work without treading on each
# other, needs a name. It gets its own ports, its own state and its own
# database, sharing only the PostgreSQL that holds them:
#
#   CHATCMS_INSTANCE=test scripts/dev-be.sh
#   CHATCMS_INSTANCE=test scripts/dev-fe.sh
#
#   CHATCMS_INSTANCE        empty for the usual instance, or small letters and digits
#   CHATCMS_HOST            address to listen on, default 127.0.0.1
#   CHATCMS_REACHABLE_HOST  address invitation links use, default the host above
#   CHATCMS_PORT            API port, default 8080, or 8081 for a named instance
#   CHATCMS_SMTP_PORT       Mailpit SMTP port, default 1025, or 1026
#   CHATCMS_INBOX_PORT      Mailpit inbox port, default 8025, or 8026
#
set -euo pipefail

cd "$(dirname "$0")/.."
root="$PWD"

# shellcheck source=scripts/dev-instance.sh
source "$root/scripts/dev-instance.sh"

reachable="${CHATCMS_REACHABLE_HOST:-$host}"
if [ "$reachable" = "0.0.0.0" ] || [ "$reachable" = "::" ]; then
    reachable="$(hostname -I 2>/dev/null | awk '{print $1}')"
    reachable="${reachable:-127.0.0.1}"
fi

if ! command -v postgres >/dev/null; then
    echo "dev-be: run this inside 'nix develop', which pins PostgreSQL, git and Hugo." >&2
    exit 1
fi

export CHATCMS_DEV_DB="$database"

if [ "${1:-}" = "stop" ]; then
    if pkill -f "$state/chatcmsd"; then
        echo "==> stopped the server"
    else
        echo "==> no server of this instance was running"
    fi

    if [ -f "$state/mailpit.pid" ] && kill "$(cat "$state/mailpit.pid")" 2>/dev/null; then
        echo "==> stopped Mailpit"
    fi
    rm -f "$state/mailpit.pid"

    if pgrep -f "\.chatcms-state[^/]*/chatcmsd" >/dev/null; then
        echo "==> leaving the database up; another instance is still running"
    else
        bash "$root/scripts/dev-db.sh" down >/dev/null 2>&1 && echo "==> stopped the database"
    fi

    sleep 1
    holder="$(ss -ltnp 2>/dev/null | grep ":$port " | grep -oP 'pid=\K[0-9]+' | head -1 || true)"
    if [ -n "$holder" ]; then
        echo "==> something still holds port $port: pid $holder ($(ps -p "$holder" -o comm= 2>/dev/null))"
        echo "    stop it with: kill $holder"
    fi
    exit 0
fi

bash "$root/scripts/dev-db.sh" up >/dev/null

if [ "${1:-}" = "fresh" ]; then
    echo "==> throwing away $database and $(basename "$state")"
    PGHOST="$root/.pgdata/sock" dropdb --if-exists "$database"
    PGHOST="$root/.pgdata/sock" createdb "$database"
    rm -rf "$state"
fi

url="postgres:///$database?host=$root/.pgdata/sock"

mkdir -p "$state"
key="$state/master.key"
if [ ! -f "$key" ]; then
    echo "==> making a master key for development"
    od -An -tx1 -N32 /dev/urandom | tr -d " \n" > "$key"
    chmod 600 "$key"
fi

if command -v mailpit >/dev/null; then
    if ! curl -sf "http://127.0.0.1:$inbox_port/api/v1/info" >/dev/null 2>&1; then
        echo "==> starting Mailpit"
        mailpit \
            --listen "$host:$inbox_port" \
            --smtp "$host:$smtp_port" \
            --database "$state/mailpit.db" \
            --quiet &
        mailpit_pid=$!
        echo "$mailpit_pid" > "$state/mailpit.pid"
    fi

    export CHATCMS_MAIL_HOST="127.0.0.1"
    export CHATCMS_MAIL_PORT="$smtp_port"
    export CHATCMS_MAIL_FROM="chatcms@localhost"
    export CHATCMS_MAIL_STARTTLS="no"
fi

export CHATCMS_DATABASE_URL="$url"
export CHATCMS_SECRET_KEY_FILE="$key"
export CHATCMS_STATE_DIR="$state"
export CHATCMS_LISTEN="$host:$port"
ai_listen="${CHATCMS_AI_LISTEN:-127.0.0.1:$ai_port}"
export CHATCMS_BASE_URL="http://$reachable:$port"

backend_said="local: previews and the assistant run as processes on this machine"
if [ "${CHATCMS_DEV_BACKEND:-local}" = "microvm" ]; then
    taps="$(ip -o link show 2>/dev/null | grep -c ': chv[0-9]*:' || true)"
    if [ "$taps" -eq 0 ]; then
        echo "dev-be: CHATCMS_DEV_BACKEND=microvm needs TAP devices; make them once with: scripts/dev-taps.sh up 2" >&2
        exit 1
    fi
    echo "==> building the workspace guest and finding its hypervisor"
    nix build .#workspace-guest --out-link "$state/workspace-guest"
    hypervisor="$(nix build --no-link --print-out-paths --inputs-from . nixpkgs#cloud-hypervisor)"
    virtiofsd="$(nix build --no-link --print-out-paths --inputs-from . nixpkgs#virtiofsd)"
    export CHATCMS_WORKSPACE_BACKEND=microvm
    export CHATCMS_WORKSPACE_GUEST="$state/workspace-guest"
    export CHATCMS_CLOUD_HYPERVISOR_BIN="$hypervisor/bin/cloud-hypervisor"
    export CHATCMS_VIRTIOFSD_BIN="$virtiofsd/bin/virtiofsd"
    export CHATCMS_WORKSPACE_TAPS="$taps"
    export CHATCMS_WORKSPACES_AT_ONCE="$taps"
    export CHATCMS_WORKSPACE_RUN_DIR="/tmp/chatcms-w${instance:+-$instance}"
    ai_listen="${CHATCMS_AI_LISTEN:-0.0.0.0:$ai_port}"
    backend_said="microvm: one micro VM per person per project, $taps at once (chv0..chv$((taps - 1)))"
fi

quarto_said="off (CHATCMS_DEV_QUARTO=0)"
if [ "${CHATCMS_DEV_QUARTO:-1}" != "0" ] && [ -z "${CHATCMS_PACKAGES:-}" ]; then
    echo "==> finding Quarto for the recipes book (the first time this downloads it)"
    if quarto="$(nix build --no-link --print-out-paths --inputs-from . nixpkgs#quarto 2>/dev/null)"; then
        export CHATCMS_PACKAGES="quarto=$quarto"
        quarto_said="$quarto"
    else
        quarto_said="could not be built, so the recipes book does not preview"
    fi
fi

export CHATCMS_AI_LISTEN="$ai_listen"

ai_env="${CHATCMS_AI_ENV:-$HOME/.config/chatcms/ai.env}"
ai_said="none; see docs/TESTING.md, Real AI inference"
if [ -r "$ai_env" ]; then
    while IFS= read -r line || [ -n "$line" ]; do
        case "$line" in
            CHATCMS_AI_*=*)
                name="${line%%=*}"
                value="${line#*=}"
                value="${value%\"}"; value="${value#\"}"
                export "$name=$value"
                ;;
        esac
    done < "$ai_env"
    ai_said="${CHATCMS_AI_MODEL:-?} via ${CHATCMS_AI_PROVIDER_URL:-?} (from $ai_env)"
fi

if [ ! -d "$state/sites" ]; then
    echo "==> seeding"
    CHATCMS_SEED_QUARTO_BOOK="$root/test/fixtures/quarto-book" go run ./cmd/seed "$url"
fi

cat <<NOTICE

==> ChatCMS on http://$reachable:$port${instance:+  (instance: $instance)}
    sign in as owner@example.org with: a long enough passphrase
    admin@example.org administers the installation, same passphrase
    marieke@korenaar.example edits korenaar, sam@korenaar.example views it,
    same passphrase; the sign-in screen has a button for each of the four

    korenaar      has a site, unpublished work, and somewhere to publish to
    studio-noord  has no site, so the setup screen has something to do
    recipes       a Quarto book, previewed as its PDF; publishing builds its website and keeps the PDF

    Mail this instance sends: http://$reachable:$inbox_port
    Workspaces: $backend_said
    Quarto: $quarto_said
    The assistant's model: $ai_said

    The frontend in another terminal, with live reloading, on :$web_port:
      ${instance_prefix}scripts/dev-fe.sh

    Ctrl-C stops this. The database keeps running; to put that away too:
      ${instance:+CHATCMS_INSTANCE=$instance }scripts/dev-be.sh stop

NOTICE

echo "==> building the preview's script and PDF viewer"
if [ -d web/node_modules ]; then
    (cd web && npx vite build --config vite.preview.config.ts --logLevel warn)
    rm -rf internal/webui/dist/__chatcms
    cp -r web/dist/__chatcms internal/webui/dist/__chatcms
else
    echo "    web/node_modules is missing, so select-to-edit and PDF previews are off; run npm install in web/"
fi

echo "==> building"
go build -o "$state/chatcmsd" ./cmd/chatcmsd

exec "$state/chatcmsd"
