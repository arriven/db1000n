# Docker VPN

## Setting up VPN for Docker users

In case of using a dedicated VPS that has banned public IP, a container with OpenVPN client can be deployed inside the same network as db1000n is in.
One of the easy ways to set it up is through the docker-compose.

`docker-compose.yml`

```yaml
{% include "../../docker-compose.yml" %}
```

`openvpn/auth.txt`:

```text
<your username for OpenVPN>
<your password for OpenVPN>
```

Also place your `*.ovpn` file into `openvpn/` directory. You can set multiple configuration files and one of them will be used.

## Start

```sh
docker-compose up -d
```

## Stop

```sh
docker-compose down
```
