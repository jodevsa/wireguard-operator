ARG WIREGUARD_GO_SRC=/usr/local/src/wireguard-go
ARG WIREGUARD_AGENT_SRC=/usr/local/src/wireguard-agent
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC=/usr/local/src/prometheus-wireguard-exporter

# step 1: Build Agent
FROM --platform=${BUILDPLATFORM} golang:1.25 AS golang-builder
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
COPY internal/iptables internal/iptables
# build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o agent -x cmd/agent/main.go
# end of step 1

# step 2: Build wireguard-go
ARG WIREGUARD_GO_SRC
WORKDIR $WIREGUARD_GO_SRC
RUN set -eux; \
    git clone https://github.com/WireGuard/wireguard-go.git . ;\
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o wireguard-go ;
# end of step 2

# step 3: Build wireguard exporter
FROM --platform=${BUILDPLATFORM} rust:1-buster AS rust-builder
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC
ARG TARGETPLATFORM

WORKDIR $PROMETHEUS_WIREGUARD_EXPORTER_SRC
RUN apt update && apt upgrade -y
RUN apt install -y g++-aarch64-linux-gnu libc6-dev-arm64-cross

ENV CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc CC_aarch64_unknown_linux_gnu=aarch64-linux-gnu-gcc CXX_aarch64_unknown_linux_gnu=aarch64-linux-gnu-g++

RUN set -eux; \
    git clone https://github.com/MindFlavor/prometheus_wireguard_exporter.git  . ;\
    if [ "$TARGETPLATFORM" = "linux/amd64" ] ; then \
            cargo build --release ;\
            cp ./target/release/prometheus_wireguard_exporter . ;\
     else if [ "$TARGETPLATFORM" = "linux/arm64" ] ; then \
            rustup target add aarch64-unknown-linux-gnu ;\
            rustup toolchain install stable-aarch64-unknown-linux-gnu ;\
            cargo build --target aarch64-unknown-linux-gnu  --release ;\
            cp ./target/aarch64-unknown-linux-gnu/release/prometheus_wireguard_exporter . ;\
    else \
            echo "$TARGETPLATFORM not supported yet." ;\
            exit 1 ;\
          fi  ;\
    fi ;
# end of step 3

# step 4: prepare image
FROM --platform=${TARGETPLATFORM} debian:bookworm
ARG WIREGUARD_GO_SRC
ARG WIREGUARD_AGENT_SRC
ARG PROMETHEUS_WIREGUARD_EXPORTER_SRC

RUN apt-get update \
    && apt-get install --no-install-recommends -y iptables wireguard-tools \
    && rm -rf /var/lib/apt/lists/*

COPY --from=golang-builder $WIREGUARD_GO_SRC/wireguard-go /usr/local/bin
COPY --from=golang-builder $WIREGUARD_AGENT_SRC/agent /usr/local/bin
COPY --from=rust-builder $PROMETHEUS_WIREGUARD_EXPORTER_SRC/prometheus_wireguard_exporter /usr/local/bin

WORKDIR /

ENTRYPOINT ["agent"]
# end of step 4
