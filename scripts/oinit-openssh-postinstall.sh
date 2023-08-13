#!/bin/sh

set -eu

# Set setuid bit on oinit-switch
chmod u+s /usr/bin/oinit-switch

# Create oinit system user
if ! getent passwd oinit >/dev/null; then
    useradd --system --shell /usr/bin/oinit-shell --home-dir /nonexistent --no-create-home --badname oinit > /dev/null
fi

# Generate host key pair
mkdir -p /etc/ssh/
! test -f /etc/ssh/host-key && ssh-keygen -t ed25519 -f /etc/ssh/host-key -N "" > /dev/null

echo "1. Please add to following two lines to '/etc/pam.d/su' in order to allow the"
echo "   oinit user to switch to other users without being prompted for a password:"
echo ""
echo "    auth [success=ignore default=1] pam_succeed_if.so use_uid user = oinit"
echo "    auth sufficient                 pam_succeed_if.so uid ne 0"

echo ""
echo ""

echo "2. Please request an OpenSSH certificate from the oinit CA administrator by"
echo "   sending him/her the file '/etc/ssh/host-key.pub'."
echo ""
echo "   You'll get two files in return:"
echo "    - host-key-cert.pub"
echo "    - user-ca.pub"
echo ""
echo "   Place them in '/etc/ssh/' and add these lines to your '/etc/ssh/sshd_config':"
echo ""
echo "    HostKey           /etc/ssh/host-key"
echo "    HostCertificate   /etc/ssh/host-key-cert.pub"
echo "    TrustedUserCAKeys /etc/ssh/user-ca.pub"
echo "    "
echo "    # You may put this at the bottom of your sshd_config file:"
echo "    Match User oinit"
echo "        PasswordAuthentication no"
