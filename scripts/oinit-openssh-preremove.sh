#!/bin/sh

set -eu

# Remove oinit system user
if getent passwd oinit >/dev/null; then
    userdel oinit > /dev/null
fi
