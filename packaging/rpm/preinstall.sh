#!/bin/sh
set -e

if ! getent group pertisk-chart >/dev/null 2>&1; then
    groupadd --system pertisk-chart
fi

if ! getent passwd pertisk-chart >/dev/null 2>&1; then
    useradd --system --gid pertisk-chart --home-dir /var/lib/pertisk-chart \
        --shell /sbin/nologin --comment "Pertisk Helm Chart Repository" pertisk-chart
fi
