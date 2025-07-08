#!/bin/sh
set -e

systemctl stop github-exporter.service || true
systemctl disable github-exporter.service || true
