#!/bin/sh

set -eu

# Remove oinit system user
if getent passwd oinit >/dev/null; then
    userdel oinit > /dev/null
fi

echo "Make sure to remove the two lines allowing 'oinit' to switch users without a"
echo "password from '/etc/pam.d/su'!"
