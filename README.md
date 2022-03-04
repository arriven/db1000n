# Death by 1000 needles

Please check existing issues (both open and closed) before creating new ones. It will save me some time answering duplicated questions and right now time is the most critical resource. Thanks.

# How to choose VPN

## Free by request or promo code
* https://clearvpn.com/ with the promo code `SAVEUKRAINE` or just click https://my.clearvpn.com/promo/redeem?code=SAVEUKRAINE you would 
  get 1 year of subscription
* NordVPN: You need to fill out form https://nordvpn.org/emergency-vpn/
* https://windscribe.com/ with promo code `ПИЗДЕЦ` you would get 30gb
* https://www.hotspotshield.com/blog/privacy-security-for-ukraine
* https://www.vpnunlimited.com/stop-russian-aggression

## Free
* https://protonvpn.com/
* https://www.f-secure.com/
* https://www.urban-vpn.com/

# HOWTO: Manual for newbies / Інструкція для новачків [ENGLISH and UKRAINIAN]

## [Українська версія] Інструкція

### Увімкніть VPN

Найкраще - це на Росію. Але інші країни також підійдуть. Головне - не залишатись у кіберпросторі України!

### Установка і запуск - для новачків

1. Качаємо свою платформу:
   - [Windows](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-windows-386.zip)
   - [Mac M1](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-darwin-arm64.tar.gz)
   - [Mac Intel](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-darwin-amd64.tar.gz)
   - [Linux 32bit](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-windows-386.zip)
   - [Linux 64bit](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-linux-amd64.tar.gz)
2. Розпаковуємо архів
3. Запускаємо файл всередині
4. Готово!

<em>Може бути застереження - “Компьютер не може підтвердити походження файлу”. Ігноруємо його, запускаємо все одно</em>

### Альтернативно: Установка і запуск - через docker

`docker run ghcr.io/arriven/db1000n`

у випадку запуску на сервері можете використовувати цю команду:
```
docker run --restart=always --name=db1000n --detach --pull=always ghcr.io/arriven/db1000n:latest
```

де:

`--restart=always` - перезапуск після помилки абощо

`--name=desinform_stop` - присвоєння імені контейнеру

`--detach` - запуск в окремому процесі

` --pull=always` - оновлення образу перед запуском

### Що робити далі

Вам потрібно лише тримати увімкненим VPN, свій ком’ютер і цю програму на ньому.
Наші спеціалісти координують атаки без відволікання вас від справ.

## [English version] Tutorial

### Use VPN!!!

Switch to Russia if possible. Don’t use Ukraine as VPN location! But any other country or VPN location is okay. Stay safe!

### For dummies

1. Download application for your platform:
   - [Windows](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-windows-386.zip)
   - [Mac M1](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-darwin-arm64.tar.gz)
   - [Mac Intel](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-darwin-amd64.tar.gz)
   - [Linux 32bit](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-windows-386.zip)
   - [Linux 64bit](https://github.com/Arriven/db1000n/releases/download/v0.5.12/db1000n-v0.5.12-linux-amd64.tar.gz)
2. Unzip the archive
3. Launch the file inside the archive
4. Done!

<em>You can get warnings from your computer about the file - ignore them. Our software is open source. It can be checked and compiled by you yourself.</em>

### Alternative: Docker usage

`docker run ghcr.io/arriven/db1000n`

or use this command in case of running on servers:
```
docker run --restart=always --name=db1000n --detach --pull=always ghcr.io/arriven/db1000n:latest
```

where:

`--restart=always` - restart after exit

`--name=desinform_stop` - assign a name to the container

`--detach` - run in separate process

` --pull=always` - pull image before running

### What’s next

You need to keep your computer active, use VPN and make sure that the application is up and running.
Our experts coordinate our attacks remotely without disturbing and interrupting you.
Thanks for your help!

## For devs

Developed by Arriven (https://github.com/Arriven)

This is a simple distributed load generation client written in go. It is able to fetch simple json config from a local or remote location. The config describes which load generation jobs should be launched in parallel. I do not intend to copy or replace it but rather provide a simple open source option. Feel free to use it in your load tests (wink-wink)

The software is provided as is under no guarantee.
I will update both the repo and this readme as I go during following days (date of writing this is 26th of February 2022, third day into russian invasion into Ukraine)

Synflood implementation is taken from https://github.com/bilalcaliskan/syn-flood and slightly patched. I couldn't just import the package as all the functionality code was in an internal package preventing import into other modules. Will figure it out better later (sorry to the owner).
## How to install

### binary install

go to releases page and install latest version for your os

### go install

run command in your terminal

```bash
go install github.com/Arriven/db1000n@latest
~/go/bin/db1000n
```

### docker install with openvpn

how to install docker?

https://docs.docker.com/get-docker/

make sure you've set all available resources to docker

https://docs.docker.com/desktop/windows/#resources
https://docs.docker.com/desktop/mac/#resources

if you don't want to use VPN from within docker container, set `--env "VPN_ENABLED=` in `run.sh`

if you want to to use VPN from within docker container:

- place your `.ovpn` or `.conf` files into `openvpn/` directory
- set `--env "VPN_ENABLED=true` in `run.sh`
- update `--env "OPENVPN_USERNAME="` and `--env "OPENVPN_PASSWORD="` in `run.sh` with your credentials

```bash
./run.sh
```

### kubernetes install

for experienced users, see instructions in [helm/](helm/)

### shell install

run install script directly into the shell (useful for install through ssh)

```bash
curl https://raw.githubusercontent.com/Arriven/db1000n/main/install.sh | sh
```

the command above will detect the os and architecture, dowload the archive, validate it, and extract db1000n executable into the working directory. You can then run it via this command

```bash
./db1000n
```

### k8s run

Enter k8s directory

```bash
$ cd k8s-manifest
```

Create deployment

```bash
$ kubectl create -f ./
```

Scale it, if you have resources

```bash
$ kubectl scale deployment/db1000n --replicas=10 -n db1000n
```

## Configuration

### Commandline reference

```text
Usage of /tmp/go-build781992389/b001/exe/main:
  -b string
        path to a backup config file in case primary one is unavailable (default "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json")
  -c string
        path to a config file, can be web endpoint (default "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json")
  -h    print help message and exit
  -l int
        logging level. 0 - Debug, 1 - Info, 2 - Warning, 3 - Error (default 1)
  -m string
        path where to dump usage metrics, can be URL or file, empty to disable
  -r duration
        refresh timeout for updating the config (default 1m0s)
```

### Config file reference

The config is expected to be in json format and has following configuration values:

- `jobs` - `[array]` array of attack job definitions to run, should be defined inside the root object
- `jobs[*]` - `[object]` single job definition as json object
- `jobs[*].type` - `[string]` type of the job (determines whhich attack function to launch). Can be `http`, `tcp`, `udp`, `syn-flood`, or `packetgen`
- `jobs[*].count` - `[number]` the amount of instances of the job to be launched, automatically set to 1 if no or invalid value is specified
- `jobs[*].args` - `[object]` arguments to pass to the job. Depends on `jobs[*].type`

`http` args:

- `method` - `[string]` http method to use (passed directly to go `http.NewRequest`)
- `path` - `[string]` url path to use (passed directly to go `http.NewRequest`)
- `body` - `[object]` http payload to use (passed directly to go `http.NewRequest`)
- `headers` - `[object]` key-value map of http headers
- `client` - `[object]` http client config for the job
- `client.tls_config` - `[object]` tls config for transport (InsecureSkipVerify is true by default)
- `client.proxy_urls` - `[array]` string urls for proxies to use (chosen randomly for each request)
- `client.timeout` - `[time.Duration]`
- `client.max_idle_connections` - `[number]`

`tcp` and `udp` shared args:

- `address` - `[string]` network host to connect to, can be either `hostname:port` or `ip:port`
- `body` - `[object]` json data to be repeatedly sent over the network

`http`, `tcp`, and `udp` shared args:

- `interval_ms` - `[number]` interval between requests in milliseconds. Defaults to 0 (Care, in case of udp job it might generate the data faster than your OS/network card can process it)
- `count` - `[number]` limit the amount of requests to send with this job invocation. Defaults to 0 (no limit). Note: if config is refreshed before this limit is reached the job will be restarted and the counter will be reset

`syn-flood` args:

- `host` - `[string]` host to attack, can be either DNS name or IP
- `port` - `[number]` port to attack
- `payload_length` - `[number]` refer to original syn-flood package docs
- `flood_type` - `[string]` type of flood to send, can be `syn`, `ack`, `synack`, and `random`

Warning: `packetgen` requires root privilleges to run

`packetgen` args:

- `host` - `[string]` host to attack, can be either DNS name or IP
- `port` - `[string]` numerical value of port to attack (string to allow template generation)
- `payload` - `[string]` payload to include into packets
- `ethernet` - `[object]` ethernet layer configuration
- `ethernet.src_mac` - `[string]`
- `ethernet.dst_mac` - `[string]`
- `ip` - `[object]` ip layer configuration
- `ip.src_ip` - `[string]`
- `ip.dst_ip` - `[string]`
- `udp` - `[object]` udp layer configuration (disables tcp if present)
- `udp.src_port` - `[number]`
- `udp.dst_port` - `[number]`
- `tcp` - `[object]` tcp layer configuration (excluded if udp is present)
- `tcp.src_port` - `[number]`
- `tcp.dst_port` - `[number]`
- `tcp.seq` - `[number]`
- `tcp.ack` - `[number]`
- `tcp.window` - `[number]`
- `tcp.urgent` - `[number]`
- `tcp.flags` - `[object]` flags for tcp (every flag has it's respective name)

Warning: `slow-loris` from testconfig.json is not yet finished and may overload the app due to not handling config refreshes

Almost every leaf `[string]` or `[object]` parameter can be templated with go template syntax. I've also added couple helper functions (list will be growing):

- `random_uuid`
- `random_int`
- `random_int_n`
- `random_ip`
- `random_payload`
- `random_mac_addr`
- `random_port`
- `local_ip`
- `local_mac_addr`
- `base64_encode`
- `base64_decode`

Please refer to official go documentation and code for these for now
