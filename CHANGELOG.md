# CHANGELOG

## v0.0.9 / 2024-04-08

* Update go toolchain for go 1.22.x
* Update go modules

## v0.0.8 / 2024-01-09

* Update go toolchain for go 1.21.x
* Update go modules

## v0.0.7 / 2023-11-20

* Update valid Identity Types from `user,device,service` to `default,router`

**NOTE**: ziti-controller v0.30.2 or newer required!

## v0.0.6 / 2023-11-08

* Force a new login when status code is 401

## v0.0.5 / 2023-11-07

* Add new login counters,
  1. `openziti_login_success_total`
  1. `openziti_login_errors_total`
* chore: Update vendoring and CI configuration.

## v0.0.4 / 2023-09-22

* Add "/fabric/v1/links" collector
* chore: Refactor exporter for consuming another API.

## v0.0.3 / 2023-09-11

* Update ziti from v0.29 to v0.30.3
* chore: Update vendoring and CI tooling.

## v0.0.2 / 2023-08-04

* Fix CI Job to run build on tags.

## v0.0.1 / 2023-08-02

* First working version.
