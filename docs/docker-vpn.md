# Setting up VPN for Docker users

In case of using a dedicated VPS that has banned public IP, a container with
OpenVPN client can be deployed inside the same network as db1000n is in.
One of the easy ways to set it up is through the docker compose:

`docker-compose.yml`:
```yaml
version: "3"

services:

  # creates OpenVPN Docker container to first provider, endpoint #1
  ovpn_01:
    image: ghcr.io/wfg/openvpn-client
    cap_add:
      - NET_ADMIN
    security_opt:
      - label:disable
    restart: unless-stopped
    volumes:
      - /dev/net:/dev/net:z
      - ./openvpn/:/data/vpn:z
    sysctls:
      - net.ipv6.conf.all.disable_ipv6=1
    environment:
      KILL_SWITCH: "on"
      HTTP_PROXY: "off"
      VPN_AUTH_SECRET: provider01_secret
      VPN_CONFIG_FILE: provider01.endpoint01.conf
    secrets:
      - provider01_secret

  # creates OpenVPN Docker container to first provider, endpoint #2
  ovpn_02:
    image: ghcr.io/wfg/openvpn-client
    cap_add:
      - NET_ADMIN
    security_opt:
      - label:disable
    restart: unless-stopped
    volumes:
      - /dev/net:/dev/net:z
      - ./openvpn/:/data/vpn:z
    sysctls:
      - net.ipv6.conf.all.disable_ipv6=1
    environment:
      KILL_SWITCH: "on"
      HTTP_PROXY: "off"
      VPN_AUTH_SECRET: provider01_secret
      VPN_CONFIG_FILE: provider01.endpoint02.conf
    secrets:
      - provider01_secret

  # creates OpenVPN Docker container to second provider, endpoint #1
  ovpn_03:
    image: ghcr.io/wfg/openvpn-client
    cap_add:
      - NET_ADMIN
    security_opt:
      - label:disable
    restart: unless-stopped
    volumes:
      - /dev/net:/dev/net:z
      - ./openvpn/:/data/vpn:z
    sysctls:
      - net.ipv6.conf.all.disable_ipv6=1
    environment:
      KILL_SWITCH: "on"
      HTTP_PROXY: "off"
      VPN_AUTH_SECRET: provider02_secret
      VPN_CONFIG_FILE: provider02.endpoint01.conf
    secrets:
      - provider02_secret

  # this Docker container will use VPN 01
  db1000n_01:
    image: ghcr.io/arriven/db1000n
    restart: unless-stopped
    depends_on:
      - ovpn_01
    network_mode: "service:ovpn_01"

  # this Docker container will use VPN 02
  db1000n_02:
    image: ghcr.io/arriven/db1000n
    restart: unless-stopped
    depends_on:
      - ovpn_02
    network_mode: "service:ovpn_02"

  # this Docker container will use VPN 03
  db1000n_03:
    image: ghcr.io/arriven/db1000n
    restart: unless-stopped
    depends_on:
      - ovpn_03
    network_mode: "service:ovpn_03"

secrets:
  provider01_secret:
    file: ./openvpn/provider01.txt
  provider02_secret:
    file: ./openvpn/provider02.txt
```

`openvpn/provider01.txt`:
```
<your username for OpenVPN provider 01>
<your password for OpenVPN provider 01>
```

`openvpn/provider02.txt`:
```
<your username for OpenVPN provider 02>
<your password for OpenVPN provider 02>
```

Also place your `provider01.endpoint01.conf`, `provider01.endpoint02.conf` and `provider02.endpoint01.conf` files into `openvpn/` directory.

## Start:

```sh
docker-compose up -d
```

## Stop:

```sh
docker-compose down
```