#!/usr/bin/env bash
set -euo pipefail

count="${2:-${CHATCMS_WORKSPACE_TAPS:-4}}"
owner="$(id -un)"

usage() {
    echo "usage: dev-taps.sh <up|down> [count]" >&2
    echo "  up    make chv0..chv<count-1> owned by ${owner}, with 10.231.<i>.1/30" >&2
    echo "  down  remove them and the firewall table" >&2
    exit 2
}

case "$count" in
    ''|*[!0-9]*) usage ;;
esac

rules() {
    cat <<'NFT'
table inet chatcms_workspaces
delete table inet chatcms_workspaces
table inet chatcms_workspaces {
  chain forward {
    type filter hook forward priority -10; policy accept;
    iifname "chv*" drop
    oifname "chv*" drop
  }
  chain input {
    type filter hook input priority -10; policy accept;
    iifname "chv*" ct state established,related accept
    iifname "chv*" tcp dport 8091 accept
    iifname "chv*" drop
  }
}
NFT
}

up() {
    local i
    for ((i = 0; i < count; i++)); do
        if ! ip link show "chv$i" >/dev/null 2>&1; then
            sudo ip tuntap add "chv$i" mode tap user "$owner"
        fi
        sudo ip addr replace "10.231.$i.1/30" dev "chv$i"
        sudo ip link set "chv$i" up
    done
    if command -v nft >/dev/null 2>&1; then
        rules | sudo nft -f -
    else
        echo "dev-taps: nft is not installed, so the VMs are not fenced off from the network" >&2
    fi
    echo "dev-taps: chv0..chv$((count - 1)) are up; set CHATCMS_WORKSPACE_TAPS=$count"
}

down() {
    local i
    if command -v nft >/dev/null 2>&1; then
        sudo nft delete table inet chatcms_workspaces 2>/dev/null || true
    fi
    for ((i = 0; i < count; i++)); do
        sudo ip link del "chv$i" 2>/dev/null || true
    done
    echo "dev-taps: removed"
}

case "${1:-}" in
    up) up ;;
    down) down ;;
    *) usage ;;
esac
