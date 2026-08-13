#!/bin/sh
set -e

if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet pertisk-chart; then
        systemctl stop pertisk-chart
    fi
    if systemctl is-enabled --quiet pertisk-chart 2>/dev/null; then
        systemctl disable pertisk-chart
    fi
fi
