# terraform-provider-tozny

Tozny Terraform provider for Infrastructure As Code (IAC) automation of Tozny products.

## Use

## Development

### Pre-requisites

* go 1.14+
* terrraform 0.13+

### Building & testing locally

```bash
make install
```

If developing on a macOS x86 environment:


```bash
make install-mac
```

## Publish

Update version number following [Semantic versioning](https://semver.org).

Build binaries of the provider for all common platforms and architectures:

```bash
make release
```

Commit and merge changes.

Distribute updated binaries to users directly or direct them to repository for instructions on how to download and use the Tozny Terraform provider.
