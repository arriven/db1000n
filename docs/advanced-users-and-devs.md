# For advanced users and developers

## For developers

*Developed by Arriven: (https://github.com/Arriven)*

This is a simple distributed load generation client written in go.
It is able to fetch simple json config from a local or remote location.
The config describes which load generation jobs should be launched in parallel.
There are other tools doing that. I do not intend to copy or replace them but rather provide a simple open source alternative so that users have more options
Feel free to use it in your load tests (wink-wink).

The software is provided as is under no guarantee.
I will update both the repo and this doc as I go during following days (date of writing this is 26th of February 2022, third day of Russian invasion into Ukraine).

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

Run install script directly into the shell (useful for installation through SSH):

```bash
curl https://raw.githubusercontent.com/Arriven/db1000n/main/install.sh | sh
```

The command above will detect the OS and architecture, download the archive, validate it, and extract `db1000n` executable into the working directory.
You can then run it via this command:

```bash
./db1000n
```

## Docker + OpenVPN

How to install docker: https://docs.docker.com/get-docker/

Make sure you've set all available resources to docker:

- https://docs.docker.com/desktop/windows/#resources
- https://docs.docker.com/desktop/mac/#resources

See [docker-vpn](docker-vpn.md) for instructions on setting it up

## Kubernetes

See [../kubernetes/](../kubernetes/) for possible ways to deploy into it

## Public Clouds

See [../terraform/](../terraform/) for possible ways to deploy into them
