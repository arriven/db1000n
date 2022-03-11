# Advanced and devs

## For developers

_Developed by [Arriven](https://github.com/Arriven)._

This is a simple distributed load generation client written in go.
It is able to fetch simple json config from a local or remote location.
The config describes which load generation jobs should be launched in parallel.
There are other tools doing that.
I do not intend to copy or replace them but rather provide a simple open source alternative so that users have more options.
Feel free to use it in your load tests (wink-wink).

The software is provided as is under no guarantee.
I will update both the repo and this doc as I go during following days (date of writing this is 26th of February 2022, third day of Russian invasion into Ukraine).

## Go installation

Run command in your terminal:

```bash
go install github.com/Arriven/db1000n@latest
~/go/bin/db1000n
```

## Shell installation

Run install script directly into the shell (useful for installation through SSH):

```bash
source <(curl https://raw.githubusercontent.com/Arriven/db1000n/main/install.sh)
```

The command above will detect the OS and architecture, download the archive, validate it, and extract `db1000n` executable into the working directory.
You can run it via this command:

```bash
./db1000n
```

## Docker + OpenVPN

How to install docker: [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/)

Make sure you've set all available resources to docker:

- [Windows](https://docs.docker.com/desktop/windows/#resources)
- [Mac](https://docs.docker.com/desktop/mac/#resources)
- [Linux](https://docs.docker.com/desktop/linux/#resources)

???+ note

    there are currently two images pointing to different configs for different usages in this repo:

    - `ghcr.io/arriven/db1000n` - default image using primitive configs that make their claim on the amount of traffic generated
    - `ghcr.io/arriven/db1000n-advanced` - image pointing to a more advanced config that has its goal in generating less traffic that is harder to detect and has more chances to actually get to the target and be processed by it (potentially exploiting known vulnerabilities). Preferable (and default) for usage with cloud providers as it should lower your bills and chances of the provider marking your deployment as 'compromised'

See [docker-vpn](docker-vpn.md) for instructions on setting it up

## Kubernetes

Here possible ways to deploy into it

- [Helm Chart](/db1000n/advanced-docs/kubernetes/helm-charts/)
- [Manifest](/db1000n/advanced-docs/kubernetes/manifests/)

## Public Clouds

See possible ways to deploy into public clouds

- [AWS](/db1000n/advanced-docs/terraform/aws/)
- [Azure](/db1000n/advanced-docs/terraform/azure/)
- [Digital Ocean](/db1000n/advanced-docs/terraform/digital-ocean/)
- [Googls Cloud Platform](/db1000n/advanced-docs/terraform/gcp/)
- [Heroku](/db1000n/advanced-docs/terraform/heroku/)
