ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/openziti_exporter /bin/openzitinode_exporter

EXPOSE      9184
USER        nobody
ENTRYPOINT  [ "/bin/openziti_exporter" ]
