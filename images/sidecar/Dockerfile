FROM alpine:latest

RUN apk --no-cache add wireguard-tools iptables

COPY start-wireguard.sh /
RUN chmod +x /start-wireguard.sh

CMD ["/start-wireguard.sh"]
