# oinit
> Certificate-based OpenSSH for Federated Identities

This repository contains a collection of programs to enable OpenSSH login for federated identities based on certificates.

Please refer to the [Documentation directory](Documentation/README.md) to learn about installation and configuration.

<p align="center">
  <img src=".github/oinit.gif" /><br>
  <i>OpenID Connect access token for selected provider is loaded from <a href="https://github.com/indigo-dc/oidc-agent">oidc-agent</a>.</i>
</p>

## Development

**Building**:

```sh
# Client application
$ make oinit

# oinit-shell and oinit-switch
$ make oinit-shell oinit-switch

# Server application (CA)
$ make oinit-ca
```

When changing the REST API annotations, run `make swagger` to generate the Swagger files.

**Testing**:

```sh
# Formating
$ make fmt

# Unit tests
$ make test

# Static analysis
$ make vet
$ make staticcheck # go install honnef.co/go/tools/cmd/staticcheck@latest
```

Alternatively, run `make all-checks` to run tests and static analysis.

### Branches

Development happens on feature branches checked out from and merged back into `prerelease`.
When ready, commits are merged into `main` and tagged as release.

[Github Actions](https://github.com/lbrocke/oinit/actions) create new Docker images for GHCR on release. The [Gitlab CI](https://codebase.helmholtz.cloud/m-team/oidc/ssh/oinit/-/pipelines) runs integration tests and creates Linux packages.

## License

This project is licensed under the [MIT License](LICENSE).
