# Docker VPN

## Setting up VPN for Docker users

In case of using a dedicated VPS that has banned public IP, a container with OpenVPN client can be deployed inside the same network as db1000n is in.
One of the easy ways to set it up is through the docker-compose.

There are few `docker-compose` examples, see `examples/docker`. Documentation you can find below:

### Static Docker Compose

`openvpn/auth.txt`:

```text
<your username for OpenVPN>
<your password for OpenVPN>
```

Also place your `*.ovpn` file into `openvpn/` directory. You can set multiple configuration files and one of them will be used.

### Old Docker Compose

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
docker-compose -f examples/docker/your_docker_file.yml up -d
```

## Stop

```sh
docker-compose -f examples/docker/your_docker_file.yml down
```
