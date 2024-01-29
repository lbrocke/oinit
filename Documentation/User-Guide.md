This is the user documentation for oinit.

## Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)

## Prerequisites

Please make sure that [oidc-agent](https://indigo-dc.gitbook.io/oidc-agent/) is installed and running. If you aren't the administrator on your machine, ask your admininistrator to install it.

## Installation

If `oinit` was already installed by your administrator, you can skip this section.

Otherwise, you have to install the `oinit` program by either building it yourself using `make oinit` or downloading it from the [latest release](https://github.com/lbrocke/oinit/releases/latest).  
Make sure the `oinit` binary is executable (`chmod +x oinit`) and resides inside a directory that is listed in your `$PATH`.

## Usage

Before connecting to an OpenSSH server for the first time, you have to tell `oinit` that it should use your federated identity instead of a password or public key:

```shell
$ oinit add login.example.com

# Wildcards are also supported
$ oinit add *.example.com

# For non-standard ports, add them via a colon
$ oinit add login.example.com:1234
```

***

After that, you should be able to use `ssh` as always, however please do not specify a username.  
Select your identity provider when asked and follow the prompts when asked for a password by `oidc-agent`.

You can optionally use environment variables to preselect a provider by url (`OIDC_ISS`/`OIDC_ISSUER`) or account name (`OIDC_AGENT_ACCOUNT`).  
If you retrieve access token from another source, you can pass it to oinit using the environment variables `ACCESS_TOKEN`, `OIDC`, `OS_ACCESS_TOKEN`, `OIDC_ACCESS_TOKEN`, `WATTS_TOKEN` or `WATTSON_TOKEN`.

```shell
$ ssh login.example.com
[1] https://aai-dev.egi.eu/auth/realms/egi
[2] https://aai.egi.eu/auth/realms/egi
[3] https://accounts.google.com
[4] https://iam.deep-hybrid-datacloud.eu
[5] https://login-dev.helmholtz.de/oauth2
[6] https://login.helmholtz.de/oauth2
[7] https://oidc.scc.kit.edu/auth/realms/kit (Accounts: kit)
[8] https://wlcg.cloud.cnaf.infn.it
? Please select a provider to use [1-8]: 7
✔ Received a certificate which is valid until 2023-07-01 14:00:00 +0200 CEST
user@host:~$
```

*You may also log in using the name of your automatically provisioned username, however it is required that you set a password beforehand.*

For `oidc-agent` version 5 and later, configurations for non-existent issuers will be created automatically. For `oidc-agent` version 4 and below, you may have to create a fitting configuration beforehand. Please refer to [the documentation](https://indigo-dc.gitbook.io/oidc-agent/user/oidc-gen) on how to do this.

Tools relying on SSH, such as scp, rsync and git should work without any further configuration.

***

You can list all hosts known to `oinit` using the `list` command. This may also include hosts that were added system-wide by your administrator.

```shell
$ oinit list
i The following hosts are managed by oinit:
	login.example.com:22
```

***

To delete a host known to oinit, you can use the `delete` command:

```shell
$ oinit delete login.example.com
✔ login.example.com was deleted.
```