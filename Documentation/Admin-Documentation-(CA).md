This is the documentation for administrators of the `oinit` certificate authority (CA).

## Contents

- [Prerequisites](#prerequisites)
- [Installation and Configuration](#installation-and-configuration)
- [Adding new OpenSSH servers](#adding-new-openssh-servers)

## Prerequisites

As the `oinit` CA doesn't support HTTPS, you should run a reverse proxy like nginx or Caddy that terminates HTTPS.

## Installation and Configuration

**1. Installation of `oinit-ca`**

An `oinit-ca` package for Debian/Ubuntu, CentOS/Fedora/Rocky Linux/Alma Linux/openSUSE is available from [repo.data.kit.edu](https://repo.data.kit.edu) (and [repo.data.kit.edu/devel/](https://repo.data.kit.edu/devel/)). You can skip steps 2, 3 and 4 when installing this package.

Alternatively, you may download the suitable `.deb`/`.rpm`/`.apk`/`.pkg.tar.zst` file from the [latest release](https://github.com/lbrocke/oinit/releases/latest). You can skip steps 2, 3 and 4 when installing this package.

A pre-built Docker image is also available, refer to section 4 (below) for this.

You can also build `oinit-ca` yourself using `make oinit-ca`.
Move the program into `/usr/local/bin/` and set the appropriate owner and permissions:

```shell
$ wget -q -O /usr/local/bin/oinit-ca https://github.com/lbrocke/oinit/releases/download/v1.0.0/oinit-ca-linux-amd64
$ chown root:root /usr/local/sbin/oinit-ca
# Make executable
$ chmod +x /usr/local/sbin/oinit-ca
```

**2. Generation of two SSH keypairs in the `/etc/oinit-ca/` directory**

The `oinit` CA needs two pairs of SSH keys, one keypair to issue certificates for hosts (host-ca) and one keypair to issued certificates for users (user-ca). Generate all four files using:

```shell
$ mkdir /etc/oinit-ca/
$ ssh-keygen -t ed25519 -f /etc/oinit-ca/user-ca -N ""
$ ssh-keygen -t ed25519 -f /etc/oinit-ca/host-ca -N ""
```

**3. Creation of configuration file**

Copy the sample configuration file from [configs/sample.config.ini](https://github.com/lbrocke/oinit/blob/main/configs/config.sample.ini) to `/etc/oinit-ca/config.ini`:

```shell
$ wget -q -O /etc/oinit-ca/config.ini https://github.com/lbrocke/oinit/blob/main/configs/config.sample.ini
```

You may have to adjust some of the default values according to the documentation in the configuration file itself.

**4. Deploy the `oinit` CA**

You can start the `oinit-ca` by providing a HTTP address to listen on, as well as the path to the configuration file.  
A sample systemd service file is provided in [init/oinit-ca.service](https://github.com/lbrocke/oinit/blob/main/init/oinit-ca.service), which you can move to `/etc/systemd/system/` and enable using `systemctl enable oinit-ca`.

```shell
$ oinit-ca 127.0.0.1:8080 /etc/oinit-ca/config.ini

# or using systemd:
$ wget -q -P /etc/systemd/system/ https://github.com/lbrocke/oinit/blob/main/init/oinit-ca.service
$ systemctl enable oinit-ca.service
$ systemctl start oinit-ca
```

If you prefer Docker, you can build an image using `make oinit-ca-docker` or download the latest release using `docker pull ghcr.io/lbrocke/oinit-ca:latest`.  
Run it by executing:

```shell
# The image expects the config file at '/etc/oinit-ca/config.ini'
$ docker run -v /etc/oinit-ca/:/etc/oinit-ca/ -p 127.0.0.1:8080:80 oinit-ca
```

You can also use the `docker-compose` template from [deploy/docker-compose.yml](https://github.com/lbrocke/oinit/blob/main/deploy/docker-compose.yml):

```shell
$ wget -q https://github.com/lbrocke/oinit/blob/main/deploy/docker-compose.yml
$ docker-compose up -d
```

**5. Configure reverse proxy**

Configure your reverse proxy to use HTTPS and send requests to `oinit` running on `http://127.0.0.1:8080`.

## Adding new OpenSSH servers

To add a new OpenSSH server, you should first request a public key (host-key.pub) as well as the public URL of the server's motley_cue instance from the OpenSSH server administrator.

Add the OpenSSH server and motley_cue address to the `/etc/oinit-ca/config.ini` file and restart the service.  
It is up to you whether you want to use an existing host-ca keypair or generate a new one for this host. Refer to *Generation of two SSH keypairs in the `/etc/oinit-ca/` directory* above on how to generate a keypair.

Lastly, you have to create and sign a new OpenSSH certificate based on the OpenSSH server public key (host-key.pub) and your host-ca private key (host-ca):

```shell
# -s: The host-ca private key used for signing
# -I: Certificate identity (text field, use server address)
# -h: This is a host certificate
# -n: OpenSSH server public DNS address
#
# You can optionally provide a validity period using -V, however keep in mind
# that you then have to re-issue this certificate regularly.
$ ssh-keygen -s /etc/oinit-ca/host-ca -I login.example.com -h -n login.example.com host-key.pub
```

Send the generated `host-key-cert.pub` file as well as the `/etc/oinit-ca/user-ca.pub` file back to the OpenSSH server administrator.
