Terraform CLO Provider
============================

Documentation: [registry.terraform.io](https://registry.terraform.io/providers/clo-ru/clo/latest/docs)

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 1.0.x
- [Go](https://golang.org/doc/install) 1.21 (to build the provider plugin)

Building The Provider
---------------------

Clone the repository

```sh
$ git clone git@github.com:clo-ru/terraform-provider-clo.git
```

Enter the provider directory and build the provider

```sh
$ cd terraform-provider-clo
$ make build
```