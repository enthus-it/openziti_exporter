ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/enthus-it/openziti_exporter"

ARG ARCH="amd64"
ARG OS="linux"
COPY openziti_exporter /bin/openziti_exporter

EXPOSE      10004
USER        nobody
ENTRYPOINT  [ "/bin/openziti_exporter" ]
