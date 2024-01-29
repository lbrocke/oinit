This is the documentation for administrators of clients that use `oinit`, e.g. pool room computers.

## Contents

- [Installation](#installation)
- [System-wide configuration](#system-wide-configuration)

## Installation

**1. oidc-agent**

Please make sure that [oidc-agent](https://indigo-dc.gitbook.io/oidc-agent/) is installed and running. If you aren't the administrator on your machine, ask your admininistrator to install it.

**2. oinit**

An `oinit` package for Debian/Ubuntu, CentOS/Fedora/Rocky Linux/Alma Linux/openSUSE is available from [repo.data.kit.edu](https://repo.data.kit.edu) (and [repo.data.kit.edu/devel/](https://repo.data.kit.edu/devel/)).

Alternatively, you may download the suitable `.deb`/`.rpm`/`.apk`/`.pkg.tar.zst` file from the [latest release](https://github.com/lbrocke/oinit/releases/latest).

You can also install the `oinit` program by building it yourself using `make oinit`.
Make sure the `oinit` binary is executable (`chmod +x oinit`) and resides inside a directory that is listed in the users's `$PATH`, e.g. `/usr/local/bin/`.

## System-wide configuration

If you want some OpenSSH servers to always use `oinit` on your clients, you can add them to the system-wide configuration file `/etc/ssh/ssh_oinit_hosts`:

```
login.example.com:22 https://ca.example.com
```

You should also make sure that the appropriate `cert-authority` line (containing the oinit CA's host-ca.pub public key) is added to OpenSSH's system-wide `/etc/ssh/ssh_known_hosts` file:

```
@cert-authority login.example.com ssh-ed25519 AAAAC3... Added by oinit
```
