This is the documentation for the administrator of an OpenSSH servers, which you want other users give access to.

## Contents

- [Prerequisites](#prerequisites)
- [Installation and Configuration](#installation-and-configuration)
- [Adding a DNS Record](#adding-a-dns-record)

## Prerequisites

Please make sure that [motley_cue](https://motley-cue.readthedocs.io/en/latest/) is installed, configured and reachable from the Internet.

## Installation and Configuration

**1. Installation of `oinit-shell` and `oinit-switch`**

An `oinit-openssh` package for Debian/Ubuntu, CentOS/Fedora/Rocky Linux/Alma Linux/openSUSE is available from [repo.data.kit.edu](https://repo.data.kit.edu) (and [repo.data.kit.edu/devel/](https://repo.data.kit.edu/devel/)). You can skip steps 2 and 3 when installing this package.

Alternatively, you may download the suitable `.deb`/`.rpm`/`.apk`/`.pkg.tar.zst` file from the [latest release](https://github.com/lbrocke/oinit/releases/latest). You can skip steps 2 and 3 when installing this package.

You can also install the `oinit` program by building it yourself using `make oinit-shell oinit-switch`.
Move the programs into `/usr/local/bin/` and set the appropriate owner and permissions:

```shell
$ wget -q -O /usr/local/bin/oinit-shell https://github.com/lbrocke/oinit/releases/download/v1.0.0/oinit-shell-linux-amd64
$ chown root:root /usr/local/bin/oinit-shell
# Make executable
$ chmod +x /usr/local/bin/oinit-shell

$ wget -q -O /usr/local/bin/oinit-switch https://github.com/lbrocke/oinit/releases/download/v1.0.0/oinit-switch-linux-amd64
$ chown root:root /usr/local/bin/oinit-switch
# Make executable
$ chmod +x /usr/local/bin/oinit-switch
```

**2. Creation of the `oinit` system user**

```shell
$ useradd --system --shell /usr/bin/oinit-shell --home-dir /nonexistent --no-create-home --badname oinit
```

**3. Generation of a SSH keypair in the `/etc/ssh/` directory**

```shell
$ ssh-keygen -t ed25519 -f /etc/ssh/host-key -N ""
```

This command will create two files:

- `/etc/ssh/host-key`     (private key)
- `/etc/ssh/host-key.pub` (public key)

**4. Request signature from CA**

Send your public key `/etc/ssh/host-key.pub` to the CA admin and ask him to sign it.  
Also tell him the public URL of your motley_cue instance, e.g. `https://login.example.com:8443`.

You'll get two files in return, move them into `/etc/ssh/` as well:

- `/etc/ssh/host-key-cert.pub`: This is your signed OpenSSH certificate based on `/etc/ssh/host-key.pub`
- `/etc/ssh/user-ca.pub`: This is a public key used to verify user certificates issued by the CA.

**5. OpenSSH server configuration**

Add the following lines to your `/etc/ssh/sshd_config` and reload your OpenSSH server afterwards:

```
HostKey			/etc/ssh/host-key
HostCertificate		/etc/ssh/host-key-cert.pub
TrustedUserCAKeys	/etc/ssh/user-ca.pub

# Optional, not strictly necessary because user 'oinit' has no password set by default.
# You may put this at the bottom of your sshd_config file, as sshd required this.
Match User oinit
	PasswordAuthentication no
```

**6. PAM configuration**

Add the following lines to `/etc/pam.d/su` to allow the oinit user to switch to other users (except root) without being prompted for a password:

```
auth [success=ignore default=1] pam_succeed_if.so use_uid user = oinit
auth sufficient                 pam_succeed_if.so uid ne 0
```

## Adding a DNS Record

To enable the automatic lookup of the oinit CA, which is responsible for your OpenSSH server, it is necessary to add a TXT record the the DNS.  
Assuming your OpenSSH server is reachable via `login.example.com`, you should create one of the TXT records, where the content is the public URL of the oinit CA.

- `_oinit-ca.login.example.com.   IN   TXT   "https://ca.example.com"` or  
- `_oinit-ca.example.com.         IN   TXT   "https://ca.example.com"`.

If you run multiple OpenSSH servers (e.g. `login{1,2,3}.example.com`), users are able to add them to their oinit configuration using wildcards (`$ oinit add *.example.com`). In this case, oinit will try to look up the TXT record with the `*` removed, resulting in `_oinit-ca.example.com.`.

If you don't want to add a DNS record, tell your users they have to manually specify the oinit CA URL in order to add your OpenSSH server to their oinit configuration:

```
$ oinit add login.example.com https://ca.example.com
```
