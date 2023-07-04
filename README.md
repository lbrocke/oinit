# oinit
> Certificate-based OpenSSH for Federated Identities

This repository contains a collection of programs to enable OpenSSH login for federated identities based on certificates.

Please refer to the [Wiki](https://github.com/lbrocke/oinit/wiki) to learn about installation and configuration.

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

## License

This project is licensed under the [MIT License](LICENSE).
