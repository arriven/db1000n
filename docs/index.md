# Quick start

## Death by 1000 needles

On 24th of February Russia has launched a full-blown invasion on Ukrainian territory. We're doing our best to stop it and prevent innocent lives being taken

!!! attention

    Please check existing issues (both open and closed) before creating new ones. It will save me some time answering duplicated questions and right now time is the most critical resource. Regards.

---

## Quickstart guide

!!! attention

    This tool is responsible only for traffic generation, you may want to use VPN if you want to test geo-blocking

### For dummies

1. Download an application for your platform:

   - [Windows](https://github.com/Arriven/db1000n/releases/latest/download/db1000n_windows_386.zip)
   - [Mac M1](https://github.com/Arriven/db1000n/releases/latest/download/db1000n_darwin_arm64.tar.gz)
   - [Mac Intel](https://github.com/Arriven/db1000n/releases/latest/download/db1000n_darwin_amd64.tar.gz)
   - [Linux 32bit](https://github.com/Arriven/db1000n/releases/latest/download/db1000n_linux_386.zip)
   - [Linux 64bit](https://github.com/Arriven/db1000n/releases/latest/download/db1000n_linux_amd64.tar.gz)

1. Unpack the archive
1. Launch the file inside the archive
1. Done!

!!! important

    Cloud providers could charge a huge amount of money not only for compute resources but for traffic as well. If you run an app in the cloud please control your billing (only advanced users are affected)!

!!! info

    You can get warnings from your computer about the file - ignore them (or allow in System Settings). Our software is open source. It can be checked and compiled by you yourself.

---

## How to install db1000n

There are different ways to install and run `db1000n`

### Binary file

Download the [latest release](https://github.com/Arriven/db1000n/releases/latest) for your arch/OS.
Unpack the archive and run it

### Docker

If you already have installed Docker, just run:

```bash
docker run --rm -it --pull always ghcr.io/arriven/db1000n
```

Or, if your container is not able to connect to your local VPN:

```bash
docker run --rm -it --pull always --network host ghcr.io/arriven/db1000n
```

### Advanced users

See [For advanced](/db1000n/advanced-docs/advanced-and-devs/)

---

## I still have questions

You will find some answers on our [FAQ](/db1000n/faq/)

---
