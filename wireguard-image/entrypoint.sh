#!/usr/bin/env bash

set -e
IPTABLE_FILE=/tmp/wireguard/iptable

function shutdown_wg() {
  echo "Shutting down Wireguard (boringtun)"
  wg-quick down "$1"
  exit 0
}

function start_wg() {
  echo "Starting up Wireguard (boringtun)"
  wg-quick up "$1"
  infinite_loop "$1"
}

function update_config() {
  echo "Update iptable rules"
  if [ -f "$IPTABLE_FILE" ]; then
    iptables-restore < "$IPTABLE_FILE"
  fi
  echo "Updating config (hot reload)"
  cp /tmp/wireguard/config /etc/wireguard/wg0.conf  
  wg syncconf wg0 <(wg-quick strip wg0)
}

function watch_and_update() {
  trap 'shutdown_wg "$1"' SIGTERM SIGINT SIGQUIT
  cp /tmp/wireguard/config /etc/wireguard/wg0.conf
  wg-quick up wg0
  if [ -f "$IPTABLE_FILE" ]; then
    iptables-restore < "$IPTABLE_FILE"
  fi
  fswatch -o /tmp/wireguard/ | (while read; do update_config; done)
}

mkdir -p /dev/net
if [ ! -c /dev/net/tun ]; then
    mknod /dev/net/tun c 10 200
fi
watch_and_update  