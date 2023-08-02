# OpenZiti exporter

[![GolangCI-Lint](https://github.com/enthus-it/openzit_exporter/actions/workflows/golangci-lint.yml/badge.svg)][golangci-lint]
[![CircleCI](https://circleci.com/gh/enthus-it/openziti_exporter/tree/main.svg?style=shield)][circleci]
[![Go Report Card](https://goreportcard.com/badge/github.com/enthus-it/openziti_exporter)][goreportcard]

Prometheus exporter for collecting [OpenZiti Management Edge API](https://openziti.io/docs/reference/developer/api/) information,
written in Go with pluggable metric collectors.

## Installation and Usage

The `openziti_exporter` listens on HTTP port 10004 by default. See the `--help` output for more options.

## Collectors

There is varying support for collectors on each operating system. The tables
below list all existing collectors and the supported systems.

Collectors are enabled by providing a `--collector.<name>` flag.
Collectors that are enabled by default can be disabled by providing a `--no-collector.<name>` flag.
To enable only some specific collector(s), use `--collector.disable-defaults --collector.<name> ...`.

### Enabled by default

Name     | Description |
---------|-------------|
identities | Exposes OpenZiti Identities from the Edge Management API. |
routers | Exposes OpenZiti Edge-Routers from the Edge Management API. |

### Disabled by default

None

## Development building and running

Prerequisites:

* [Go compiler](https://golang.org/dl/)
* Access to the [OpenZiti Edge Management API](https://openziti.io/docs/reference/developer/api/)

Building:

```shell
    git clone https://github.com/enthus-it/openziti_exporter.git
    cd openziti_exporter
    make build
    ./openziti_exporter <flags>
```

To see all available configuration flags:

```shell
    ./openziti_exporter -h
```

## Running tests

```shell
    make test
```

## TLS endpoint

**EXPERIMENTAL** The exporter supports TLS via a new web configuration file.

```shell
    ./openziti_exporter --web.config.file=web-config.yml
```

See the [exporter-toolkit https package](https://github.com/prometheus/exporter-toolkit/blob/v0.1.0/https/README.md) for more details.

[golangci-lint]: https://github.com/enthus-it/openziti_exporter/actions/workflows/golangci-lint.yml
[circleci]: https://circleci.com/gh/enthus-it/openziti_exporter
[goreportcard]: https://goreportcard.com/report/github.com/enthus-it/openziti_exporter
