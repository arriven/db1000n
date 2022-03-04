#!/bin/sh

# wrapper script

# run normal up.sh script that sets DNS, etc
/etc/openvpn/up.sh
touch /tmp/tunnel