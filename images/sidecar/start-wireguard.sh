#!/bin/sh

set -e

# Generate WireGuard keys
umask 077

# Start WireGuard
wg-quick up wg0

# Configure iptables to route traffic over the VPN
iptables -A FORWARD -i eth0 -o wg0 -j ACCEPT
iptables -A FORWARD -i wg0 -o eth0 -j ACCEPT
iptables -t nat -A POSTROUTING -o wg0 -j MASQUERADE

# Start the main container process
exec "$@"
