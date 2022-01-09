ARG BORINGTUN_SRC=/usr/local/src/boringtun

FROM rust:1-buster AS builder
ARG BORINGTUN_VERSION=master
ARG BORINGTUN_SRC

WORKDIR $BORINGTUN_SRC

RUN set -eux; \
    git clone -b "${BORINGTUN_VERSION}" --depth=1 \
    "https://github.com/cloudflare/boringtun.git" . ;\
    RUSTFLAGS="${RUSTFLAGS:-} -A unused_must_use" cargo build --release; \
    strip ./target/release/boringtun

FROM alpine:3.14
ARG BORINGTUN_SRC

ENV WG_QUICK_USERSPACE_IMPLEMENTATION=boringtun
ENV WG_THREADS=2
ENV WG_SUDO=1
ENV WG_LOG_LEVEL=info
ENV WG_LOG_FILE=/dev/stdout
ENV WG_ERR_LOG_FILE=/dev/stderr
ENV SUB_NET="10.8.0.0/24"

RUN apk add wireguard-tools iproute2 iptables gettext

COPY --from=builder $BORINGTUN_SRC/target/release/boringtun /usr/local/bin
COPY entrypoint.sh /
COPY wg0.conf.template /



WORKDIR /etc/wireguard

ENTRYPOINT [ "/entrypoint.sh"]