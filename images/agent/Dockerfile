ARG WIREGUARD_GO_SRC=/usr/local/src/wireguard-go
ARG WIREGUARD_AGENT_SRC=/usr/local/src/wireguard-agent
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC=/usr/local/src/prometheus-wireguard-exporter


FROM golang:1-buster AS golang-builder
ARG WIREGUARD_AGENT_SRC
WORKDIR $WIREGUARD_AGENT_SRC
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# COPY the Go code
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY internal internal/

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download



RUN go build -o agent cmd/agent/main.go

ARG WIREGUARD_GO_SRC

WORKDIR $WIREGUARD_GO_SRC

RUN set -eux; \
    git clone https://github.com/WireGuard/wireguard-go.git .;\
    make; \
    strip ./wireguard-go

FROM rust:1-buster AS builder
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC
WORKDIR $PROMETHEUS_WIREGUARD_EXPORTER_SRC

RUN set -eux; \
    git clone https://github.com/MindFlavor/prometheus_wireguard_exporter.git  .;\
    RUSTFLAGS="${RUSTFLAGS:-} -A unused_must_use" cargo build --release; \
    strip ./target/release/prometheus_wireguard_exporter


   
FROM debian:buster-slim
ARG WIREGUARD_GO_SRC
ARG WIREGUARD_AGENT_SRC
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC

ENV WG_QUICK_USERSPACE_IMPLEMENTATION=wireguard-go

RUN set -eux; \
    echo 'deb http://deb.debian.org/debian buster-backports main' > /etc/apt/sources.list.d/backports.list; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-suggests --no-install-recommends \
    wireguard-tools sudo iproute2 iptables gettext-base fswatch

COPY --from=golang-builder $WIREGUARD_GO_SRC/wireguard-go /usr/local/bin
COPY --from=golang-builder $WIREGUARD_AGENT_SRC/agent /usr/local/bin
COPY --from=builder $PROMETHEUS_WIREGUARD_EXPORTER_SRC/target/release/prometheus_wireguard_exporter /usr/local/bin



WORKDIR /etc/wireguard

ENTRYPOINT ["agent"]
