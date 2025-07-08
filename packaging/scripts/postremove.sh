#!/bin/sh
set -e

if [ ! -d /var/lib/github-exporter ]; then
    userdel github-exporter 2>/dev/null || true
    groupdel github-exporter 2>/dev/null || true
fi
