# oinit
> Certificate-based OpenSSH for Federated Identities

[![GitHub release](https://img.shields.io/github/release/lbrocke/oinit?include_prereleases=&sort=semver&color=blue)](https://github.com/lbrocke/oinit/releases/)
[![License](https://img.shields.io/badge/license-MIT-blue)](https://github.com/lbrocke/oinit/blob/main/LICENSE)
[![Gitlab CI](https://git.scc.kit.edu/m-team/oidc/ssh/oinit/badges/main/pipeline.svg)](https://git.scc.kit.edu/m-team/oidc/ssh/oinit/-/pipelines)

This repository contains a collection of programs to enable OpenSSH login for federated identities based on certificates.

Please refer to the [Wiki](https://github.com/lbrocke/oinit/wiki) to learn about installation and configuration.

<p align="center">
  <img src=".github/oinit.gif" /><br>
  <i>OpenID Connect access token for selected provider is loaded from <a href="https://github.com/indigo-dc/oidc-agent">oidc-agent</a>.</i>
</p>

## Development

```sh
# Client application
$ make oinit

# oinit-shell and oinit-switch
$ make oinit-shell oinit-switch

# Server application (CA)
$ make oinit-ca
```

When changing the REST API annotations, run `make swagger` to generate the Swagger files.

### Branches

Development happens on feature branches checked out from and merged back into `prerelease`.
When ready, commits are merged into `main` and tagged as release.

[Github Actions](https://github.com/lbrocke/oinit/actions) create new Docker images for GHCR on release. The [Gitlab CI](https://git.scc.kit.edu/m-team/oinit/-/pipelines) runs integration tests and creates Linux packages.

## License

This project is licensed under the [MIT License](LICENSE).
