# oinit
> Certificate-based single sign on for OpenSSH

This repository contains a collection of programs to enable single sign-on in OpenSSH.

## Building

```sh
# Client application
$ make oinit

# Server application (CA)
$ make oinit-ca
```

## Development

When changing the REST API annotations, run `make swagger` to generate the Swagger files.
