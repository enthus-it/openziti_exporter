ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="Mario Trangoni <mjtrangoni@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/enthus-it/openziti_exporter"
LABEL org.opencontainers.image.description="Prometheus exporter for collecting OpenZiti Management Edge API information"
LABEL org.opencontainers.image.licenses=Apache-2.0

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/openziti_exporter /bin/openziti_exporter

EXPOSE      10004
USER        nobody
ENTRYPOINT  [ "/bin/openziti_exporter" ]
