# Setting up VPN for Docker users

In case of using a dedicated VPS that has banned public IP, a container with
OpenVPN client can be deployed inside the same network as db1000n is in.
One of the easy ways to set it up is through the docker compose:

`docker-compose.yml`:
```yaml
version: "3"
services:

  ovpn:
    image: dperson/openvpn-client
    cap_add:
      - NET_ADMIN
    security_opt:
      - label:disable
    # read_only: true
    restart: unless-stopped
    volumes:
      - /dev/net:/dev/net:z
      - /path/to/dir/with/certificate:/vpn
    # NOTE: user/password can be defined:
    # 1. In .ovpn configuration file. In this case put it inside
    #    /path/to/dir/with/certificate (see above)
    # 2. In the `command` directive (see below).
    #    If you are using .ovpn file, then comment out three lines below
    command:
      - "-v"
      - "host;user;password"

  db1000n:
    image: ghcr.io/arriven/db1000n
    restart: unless-stopped
    depends_on:
      - ovpn
    network_mode: "service:ovpn"
```


## Run:

```sh
docker compose up -d
```


You can also run multiple db1000n instances with different VPN
configurations (i.e. multiple countries):

```yaml
version: "3"
services:
  ovpn1:
    image: dperson/openvpn-client
    # ...
    volumes:
      - /certificate1:/vpn
  db1000n1:
    image: ghcr.io/arriven/db1000n
    # ...
    network_mode: "service:ovpn1"
  ovpn2:
    image: dperson/openvpn-client
    # ...
    volumes:
      - /certificate2:/vpn
  db1000n1:
    image: ghcr.io/arriven/db1000n
    # ...
    network_mode: "service:ovpn2"
  # more services...
```
