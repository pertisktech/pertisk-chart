#!/bin/sh
# Launcher for pertisk-chart. Reads /etc/pertisk-chart/config.conf.
set -e

if [ -f /etc/pertisk-chart/config.conf ]; then
    # shellcheck source=/dev/null
    . /etc/pertisk-chart/config.conf
fi

set -- \
    --port="${PORT:-7080}" \
    --data-dir="${DATA_DIR:-/var/lib/pertisk-chart}" \
    --storage-local-rootdir="${STORAGE_ROOT:-/var/lib/pertisk-chart/chartstorage}" \
    --db-type="${DB_TYPE:-sqlite}" \
    --web-dir="${WEB_DIR:-/usr/share/pertisk-chart/web}"

[ -n "${DB_DSN:-}" ] && set -- "$@" --db-dsn="$DB_DSN"
[ -n "${JWT_SECRET:-}" ] && set -- "$@" --jwt-secret="$JWT_SECRET"
[ "${ENABLE_METRICS:-}" = "true" ] && set -- "$@" --enable-metrics
[ "${DEBUG:-}" = "true" ] && set -- "$@" --debug
[ "${ENABLE_HTTP3:-}" = "true" ] && set -- "$@" --enable-http3
[ -n "${TLS_CERT:-}" ] && set -- "$@" --tls-cert="$TLS_CERT"
[ -n "${TLS_KEY:-}" ] && set -- "$@" --tls-key="$TLS_KEY"
[ -n "${ENABLE_ZSTD:-}" ] && set -- "$@" --enable-zstd="$ENABLE_ZSTD"

exec /usr/bin/pertisk-chart "$@"
