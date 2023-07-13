#!/bin/sh

set -eu

if command -v systemctl > /dev/null && [ "$(systemctl is-system-running)" != "offline" ]; then
    # load new oinit-ca.service file
    systemctl daemon-reload
fi

mkdir -p /etc/oinit-ca/
ssh-keygen -t ed25519 -f /etc/oinit-ca/user-ca -N "" > /dev/null
ssh-keygen -t ed25519 -f /etc/oinit-ca/host-ca -N "" > /dev/null
