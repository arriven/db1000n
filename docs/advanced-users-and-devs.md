# For advanced users and developers

## For developers

*Developed by Arriven: (https://github.com/Arriven)*

This is a simple distributed load generation client written in go.
It is able to fetch simple json config from a local or remote location.
The config describes which load generation jobs should be launched in parallel.
I do not intend to copy or replace it but rather provide a simple open source option.
Feel free to use it in your load tests (wink-wink).

The software is provided as is under no guarantee.
I will update both the repo and this doc as I go during following days (date of writing this is 26th of February 2022, third day into russian invasion into Ukraine),

Synflood implementation is taken from [bilalcaliskan/syn-flood](https://github.com/bilalcaliskan/syn-flood) and slightly patched.
I couldn't just import the package as all the functionality code was in an internal package preventing import into other modules.
Will figure it out better later (sorry to the owner)

## Go installation

Run command in your terminal:

```bash
go install github.com/Arriven/db1000n@latest
~/go/bin/db1000n
```

## Shell installation

Run install script directly into the shell (useful for install through ssh):

```bash
curl https://raw.githubusercontent.com/Arriven/db1000n/main/install.sh | sh
```

The command above will detect the os and architecture, dowload the archive, validate it, and extract `db1000n` executable into the working directory.
You can then run it via this command:

```bash
./db1000n
```

## Docker + OpenVPN

How to install docker: https://docs.docker.com/get-docker/

Make sure you've set all available resources to docker:

- https://docs.docker.com/desktop/windows/#resources
- https://docs.docker.com/desktop/mac/#resources

If you don't want to use VPN from within docker container, set `--env "VPN_ENABLED=false"` in `run.sh`

If you want to to use VPN from within docker container:

- Place your `.ovpn` or `.conf` files into `openvpn/` directory
- If there's more than one `.ovpn` or `.conf` file, random configuration will be picked
- Set `--env "VPN_ENABLED=true"` in `run.sh`
- Update `--env "OPENVPN_USERNAME="` and `--env "OPENVPN_PASSWORD="` in `run.sh` with your credentials

```bash
./run.sh
```

If you want to use different VPN providers (meaning different `OPENVPN_USERNAME` and `OPENVPN_PASSWORD` credentials), do the following:

- Set `--env "VPN_ENABLED=true"` in `run.sh`
- Remove `--env "OPENVPN_USERNAME="` and `--env "OPENVPN_PASSWORD="` from `run.sh`
- Place your `openvpn-provider-1.conf` into `openvpn/` directory
- Find the line in your `openvpn-provider-1.conf` where it says `auth-user-pass` and replace with `auth-user-pass provider1.txt`
- Create a text file in `openvpn/provider1.txt` with two lines:

```
<your provider 1 username>
<your provider 1 password>
```

- Repeat steps above for multiple providers

## Kubernetes

See [kubernetes/](kubernetes/) for possible ways to deploy into it

## Public Clouds

See [terraform/](terraform/) for possible ways to deploy into them
