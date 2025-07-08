#!/bin/sh
set -e

if ! getent group github-exporter >/dev/null 2>&1; then
    groupadd --system github-exporter
fi

if ! getent passwd github-exporter >/dev/null 2>&1; then
    useradd --system --create-home --home-dir /var/lib/github-exporter --shell /bin/bash -g github-exporter github-exporter
fi
