ARG BORINGTUN_SRC=/usr/local/src/boringtun
ARG WIREGUARD_GO_SRC=/usr/local/src/wireguard-go
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC=/usr/local/src/prometheus-wireguard-exporter


FROM golang:1-buster AS golang-builder
ARG WIREGUARD_GO_SRC

WORKDIR $WIREGUARD_GO_SRC

RUN set -eux; \
    git clone https://github.com/WireGuard/wireguard-go.git .;\
    make; \
    strip ./wireguard-go

FROM rust:1-buster AS builder
ARG BORINGTUN_VERSION=master
ARG BORINGTUN_SRC
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC

#WORKDIR $BORINGTUN_SRC

#RUN set -eux; \
#    git clone -b "${BORINGTUN_VERSION}" --depth=1 \
#    "https://github.com/cloudflare/boringtun.git" . ;\
#    RUSTFLAGS="${RUSTFLAGS:-} -A unused_must_use" cargo build --release; \
#    strip ./target/release/boringtun



WORKDIR $PROMETHEUS_WIREGUARD_EXPORTER_SRC

RUN set -eux; \
    git clone https://github.com/MindFlavor/prometheus_wireguard_exporter.git  .;\
    RUSTFLAGS="${RUSTFLAGS:-} -A unused_must_use" cargo build --release; \
    strip ./target/release/prometheus_wireguard_exporter


   
FROM debian:buster-slim
ARG BORINGTUN_SRC
ARG WIREGUARD_GO_SRC
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC

ENV WG_QUICK_USERSPACE_IMPLEMENTATION=wireguard-go
ENV WG_THREADS=4
ENV WG_SUDO=1
ENV WG_LOG_LEVEL=info
ENV WG_LOG_FILE=/dev/stdout
ENV WG_ERR_LOG_FILE=/dev/stderr
ENV SUB_NET="10.8.0.0/24"

RUN set -eux; \
    echo 'deb http://deb.debian.org/debian buster-backports main' > /etc/apt/sources.list.d/backports.list; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-suggests --no-install-recommends \
    wireguard-tools sudo iproute2 iptables gettext-base fswatch

COPY --from=golang-builder $WIREGUARD_GO_SRC/wireguard-go /usr/local/bin
#COPY --from=builder $BORINGTUN_SRC/target/release/boringtun /usr/local/bin
COPY --from=builder $PROMETHEUS_WIREGUARD_EXPORTER_SRC/target/release/prometheus_wireguard_exporter /usr/local/bin
COPY entrypoint.sh /


WORKDIR /etc/wireguard

ENTRYPOINT [ "/entrypoint.sh"]
