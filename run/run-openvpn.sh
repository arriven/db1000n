#!/bin/sh

set -e

if [[ -n "$VPN_ENABLED" ]]
then
  # prepare vpn
  mkdir -p /dev/net
  if [ ! -c /dev/net/tun ]; then
  	mknod /dev/net/tun c 10 200
  fi

  readonly OVPNCONF_DIR="/etc/openvpn_host"

  # Move the config over to a tmp file and inject the credentials in
  OPENVPN_CONF="$(mktemp).ovpn"
  readonly OPENVPN_CONF

  if [[ -n "$OPENVPN_USERNAME" ]] || [[ -n "$OPENVPN_PASSWORD" ]]
  then
  	readonly OPENVPN_AUTH_USER_PASS="$(mktemp)_credentials.conf"
  	echo -e "$OPENVPN_USERNAME\n$OPENVPN_PASSWORD" > "$OPENVPN_AUTH_USER_PASS"
  	chmod 400 $OPENVPN_AUTH_USER_PASS

  	grep -Ev "(^(up|down)\s|auth-user-pass)" \
  	  "$(ls $OVPNCONF_DIR/*.conf $OVPNCONF_DIR/*.ovpn | shuf -n 1)" > "$OPENVPN_CONF"
  	echo "auth-user-pass $OPENVPN_AUTH_USER_PASS" >> "$OPENVPN_CONF"
  else
  	grep -Ev "(^(up|down)\s)" \
  	  "$(ls $OVPNCONF_DIR/*.conf $OVPNCONF_DIR/*.ovpn | shuf -n 1)" > "$OPENVPN_CONF"
  fi

  /usr/sbin/openvpn --script-security 2 --up /usr/local/bin/openvpn-up.sh \
    --cd /etc/openvpn_host --config "$OPENVPN_CONF"

  # if openvpn exits, take the whole container with us
  echo openvpn died
  kill 1 # kill the supervisor so the container dies
else
  touch /tmp/tunnel
fi
