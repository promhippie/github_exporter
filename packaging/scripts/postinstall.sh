#!/bin/sh
set -e

chown -R github-exporter:github-exporter /var/lib/github-exporter
chmod 750 /var/lib/github-exporter

if [ -d /run/systemd/system ]; then
    systemctl daemon-reload

    if systemctl is-enabled --quiet github-exporter.service; then
        systemctl restart github-exporter.service
    fi
fi
