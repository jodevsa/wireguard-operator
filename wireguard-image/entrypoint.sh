#!/usr/bin/env bash

set -e


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
function setup_NAT() {
  echo "Setting up NAT"
  /usr/sbin/iptables-legacy -t nat -I POSTROUTING 1 -s $SUB_NET -o eth0 -j MASQUERADE
}
function update_config() {
  echo "Updating config (hot reload)"
  cp /tmp/wireguard/config /etc/wireguard/wg0.conf  
  wg syncconf wg0 <(wg-quick strip wg0)
}

function watch_and_update() {
  trap 'shutdown_wg "$1"' SIGTERM SIGINT SIGQUIT
  cp /tmp/wireguard/config /etc/wireguard/wg0.conf
  wg-quick up wg0
  fswatch -o /tmp/wireguard/ | (while read; do update_config; done)
}

mkdir -p /dev/net
if [ ! -c /dev/net/tun ]; then
    mknod /dev/net/tun c 10 200
fi
setup_NAT
watch_and_update  