#!/bin/sh
set -e

chown -R pertisk-chart:pertisk-chart /var/lib/pertisk-chart
chown -R pertisk-chart:pertisk-chart /var/log/pertisk-chart
chmod 750 /var/lib/pertisk-chart
chmod 750 /var/log/pertisk-chart

mkdir -p /var/lib/pertisk-chart/chartstorage
chown pertisk-chart:pertisk-chart /var/lib/pertisk-chart/chartstorage
chmod 750 /var/lib/pertisk-chart/chartstorage

if [ -d /usr/share/pertisk-chart/web ]; then
    chown -R root:root /usr/share/pertisk-chart/web
    chmod -R 755 /usr/share/pertisk-chart/web
fi

if command -v setcap >/dev/null 2>&1; then
    setcap 'cap_net_bind_service=+ep' /usr/bin/pertisk-chart
else
    echo "Warning: setcap not found. Install libcap, then run:"
    echo "  sudo setcap 'cap_net_bind_service=+ep' /usr/bin/pertisk-chart"
fi

if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload
    echo ""
    echo "Pertisk Chart Repository has been installed successfully!"
    echo ""
    echo "IMPORTANT: Create an admin user before starting the service:"
    echo "  sudo pertisk-chart-create-admin \\"
    echo "    -username admin \\"
    echo "    -email admin@example.com \\"
    echo "    -password your-secure-password \\"
    echo "    -data-dir /var/lib/pertisk-chart \\"
    echo "    -db-type sqlite"
    echo ""
    echo "Then start the service:"
    echo "  sudo systemctl start pertisk-chart"
    echo "  sudo systemctl enable pertisk-chart"
    echo ""
    echo "Configuration: /etc/pertisk-chart/config.conf"
fi
