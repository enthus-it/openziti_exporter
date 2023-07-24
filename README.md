# OpenZiti exporter

[![Go Report Card](https://goreportcard.com/badge/github.com/enthus-it/openziti_exporter)][goreportcard]

Prometheus exporter for collecting [OpenZiti Management Edge API](https://openziti.io/docs/reference/developer/api/) information,
written in Go with pluggable metric collectors.

## Installation and Usage

The `openziti_exporter` listens on HTTP port 9184 by default. See the `--help` output for more options.

## Collectors

There is varying support for collectors on each operating system. The tables
below list all existing collectors and the supported systems.

Collectors are enabled by providing a `--collector.<name>` flag.
Collectors that are enabled by default can be disabled by providing a `--no-collector.<name>` flag.
To enable only some specific collector(s), use `--collector.disable-defaults --collector.<name> ...`.

### Include & Exclude flags

A few collectors can be configured to include or exclude certain patterns using dedicated flags. The exclude flags are used to indicate "all except", while the include flags are used to say "none except". Note that these flags are mutually exclusive on collectors that support both.

Example:

```txt
--collector.filesystem.mount-points-exclude=^/(dev|proc|sys|var/lib/docker/.+|var/lib/kubelet/.+)($|/)
```

List:

Collector | Scope | Include Flag | Exclude Flag
--- | --- | --- | ---
identities | management | --collector.arp.device-include | --collector.arp.device-exclude

### Enabled by default

Name     | Description | OS
---------|-------------|----
identities | Exposes OpenZiti Identities from the Edge Management API. | Any

### Disabled by default

None

## Development building and running

Prerequisites:

* [Go compiler](https://golang.org/dl/)
* Access to the [OpenZiti Edge Management API](https://openziti.io/docs/reference/developer/api/)

Building:

    git clone https://github.com/enthus-it/openziti_exporter.git
    cd openziti_exporter
    make build
    ./openziti_exporter <flags>

To see all available configuration flags:

    ./openziti_exporter -h

## Running tests

    make test

## TLS endpoint

** EXPERIMENTAL **

The exporter supports TLS via a new web configuration file.

```console
./openziti_exporter --web.config.file=web-config.yml
```

See the [exporter-toolkit https package](https://github.com/prometheus/exporter-toolkit/blob/v0.1.0/https/README.md) for more details.

[goreportcard]: https://goreportcard.com/report/github.com/enthus-it/openziti_exporter
