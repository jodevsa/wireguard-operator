ARG GO_VERSION=1.20
ARG TARGETOS
ARG TARGETARCH
ARG WIREGUARD_GO_SRC=/usr/local/src/wireguard-go
ARG WIREGUARD_AGENT_SRC=/usr/local/src/wireguard-agent
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC=/usr/local/src/prometheus-wireguard-exporter

# step 1: Build Agent
FROM --platform=${BUILDPLATFORM} golang:1.20 AS golang-builder
ARG WIREGUARD_AGENT_SRC
WORKDIR $WIREGUARD_AGENT_SRC
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
# COPY src
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY internal internal/
# build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o agent -x cmd/agent/main.go

# step 2: Build wireguard-go
ARG WIREGUARD_GO_SRC
WORKDIR $WIREGUARD_GO_SRC
RUN set -eux; \
    git clone https://github.com/WireGuard/wireguard-go.git .;\
    make; \
    strip ./wireguard-go

# step 3: Build wireguard exporter
FROM --platform=${BUILDPLATFORM} rust:1-buster AS rust-builder
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC
WORKDIR $PROMETHEUS_WIREGUARD_EXPORTER_SRC

RUN set -eux; \
    git clone https://github.com/MindFlavor/prometheus_wireguard_exporter.git  .;\
    RUSTFLAGS="${RUSTFLAGS:-} -A unused_must_use" cargo build --release; \
    strip ./target/release/prometheus_wireguard_exporter


# step 4: prepare image
FROM --platform=${BUILDPLATFORM} debian:buster-slim
ARG WIREGUARD_GO_SRC
ARG WIREGUARD_AGENT_SRC
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC

RUN set -eux; \
    echo 'deb http://deb.debian.org/debian buster-backports main' > /etc/apt/sources.list.d/backports.list; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-suggests --no-install-recommends \
    wireguard-tools sudo iproute2 iptables gettext-base fswatch

COPY --from=golang-builder $WIREGUARD_GO_SRC/wireguard-go /usr/local/bin
COPY --from=golang-builder $WIREGUARD_AGENT_SRC/agent /usr/local/bin
COPY --from=rust-builder $PROMETHEUS_WIREGUARD_EXPORTER_SRC/target/release/prometheus_wireguard_exporter /usr/local/bin

WORKDIR /etc/wireguard

ENTRYPOINT ["agent"]
