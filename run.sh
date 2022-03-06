#!/usr/bin/env bash

set -euo pipefail

PROJECT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
readonly PROJECT_DIR
readonly OPENVPN_DIR="$PROJECT_DIR/openvpn"

# Build the docker image and tag it
docker build -t db1000n-openvpn --target openvpn "$PROJECT_DIR"
docker run \
    --rm \
    --cap-add=NET_ADMIN \
    --mount type=bind,source="$OPENVPN_DIR",target=/etc/openvpn_host,readonly \
    --sysctl net.ipv6.conf.all.disable_ipv6=0 \
    --env "OPENVPN_USERNAME=" \
    --env "OPENVPN_PASSWORD=" \
    --env "VPN_ENABLED=true" \
    db1000n-openvpn
