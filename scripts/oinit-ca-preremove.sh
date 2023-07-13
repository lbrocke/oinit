#!/bin/sh

set -eu

if command -v systemctl >/dev/null; then
    systemctl stop oinit-ca || true
fi
