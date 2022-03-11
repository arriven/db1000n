# Docker VPN

## Setting up VPN for Docker users

In case of using a dedicated VPS that has banned public IP, a container with OpenVPN client can be deployed inside the same network as db1000n is in.
One of the easy ways to set it up is through the docker-compose.

`docker-compose.yml`

```yaml
{% include "../../docker-compose.yml" %}
```

`openvpn/provider01.txt`:

```text
<your username for OpenVPN provider 01>
<your password for OpenVPN provider 01>
```

`openvpn/provider02.txt`:

```text
<your username for OpenVPN provider 02>
<your password for OpenVPN provider 02>
```

Also place your `provider01.endpoint01.conf`, `provider01.endpoint02.conf` and `provider02.endpoint01.conf` files into `openvpn/` directory.

## Start

```sh
docker-compose up -d
```

## Stop

```sh
docker-compose down
```
