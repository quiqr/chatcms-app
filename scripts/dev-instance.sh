# shellcheck shell=bash disable=SC2034,SC2154
instance="${CHATCMS_INSTANCE:-}"
if [ -n "$instance" ]; then
    case "$instance" in
        *[!a-z0-9]*)
            echo "dev: an instance is named with small letters and digits, so that it can name a database too." >&2
            exit 1
            ;;
    esac
    state="$root/.chatcms-state-$instance"
    database="chatcms_dev_$instance"
    port="${CHATCMS_PORT:-8081}"
    web_port="${CHATCMS_WEB_PORT:-5174}"
    smtp_port="${CHATCMS_SMTP_PORT:-1026}"
    inbox_port="${CHATCMS_INBOX_PORT:-8026}"
    ai_port="${CHATCMS_AI_PORT:-8092}"
else
    state="$root/.chatcms-state"
    database="chatcms_dev"
    port="${CHATCMS_PORT:-8080}"
    web_port="${CHATCMS_WEB_PORT:-5173}"
    smtp_port="${CHATCMS_SMTP_PORT:-1025}"
    inbox_port="${CHATCMS_INBOX_PORT:-8025}"
    ai_port="${CHATCMS_AI_PORT:-8091}"
fi

host="${CHATCMS_HOST:-127.0.0.1}"

instance_prefix="${instance:+CHATCMS_INSTANCE=$instance }"
if [ "$host" != "127.0.0.1" ] && [ "$host" != "localhost" ]; then
    instance_prefix="${instance_prefix}CHATCMS_HOST=$host "
fi
