# Configuration

## Command Line reference

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

## Config file reference

The config is expected to be in json format and has following configuration values:

- `jobs` - `[array]` array of attack job definitions to run, should be defined inside the root object
- `jobs[*]` - `[object]` single job definition as json object
- `jobs[*].type` - `[string]` type of the job (determines which attack function to launch). Can be `http`, `tcp`, `udp`, `syn-flood`, or `packetgen`
- `jobs[*].count` - `[number]` the amount of instances of the job to be launched, automatically set to 1 if no or invalid value is specified
- `jobs[*].args` - `[object]` arguments to pass to the job. Depends on `jobs[*].type`

`http` args:

- `request` - `[object]` defines requests to be sent
- `request.method` - `[string]` http method to use (passed directly to go `http.NewRequest`)
- `request.path` - `[string]` url path to use (passed directly to go `http.NewRequest`)
- `request.body` - `[object]` http payload to use (passed directly to go `http.NewRequest`)
- `request.headers` - `[object]` key-value map of http headers
- `request.cookies` - `[object]` key-value map of http cookies (you can still set cookies directly via the header with `cookie_string` template function or statically, see `config/examples/advanced/ddos-guard.yaml` for an example)
- `client` - `[object]` http client config for the job
- `client.tls_config` - `[object]` tls config for transport (InsecureSkipVerify is true by default)
- `client.proxy_urls` - `[array]` comma-separated list of string urls for proxies to use (chosen randomly for each request)
- `client.timeout` - `[time.Duration]`
- `client.max_idle_connections` - `[number]`

`tcp` args:

- `proxy_urls` - `[string]` comma-separated list of http/socks5 proxies to use (chosen randomly for each job)
- `timeout` - `[duration]` timeout for connecting to the proxy

`tcp` and `udp` shared args:

- `address` - `[string]` network host to connect to, can be either `hostname:port` or `ip:port`
- `body` - `[object]` json data to be repeatedly sent over the network

Warning: `packetgen` requires root privileges to run

`packetgen` args:

- `connection` - `[object]` raw ip connection parameters
- `connection.name` - `[string]` name of the network to use. can be `ip4:tcp`, `ip6:tcp`, `ip4:udp`, `ip6:udp`, or anything else supported by the go runtime
- `connection.address` - `[string]` address of the interface used to send packets (on the attacking machine)
- `packet` - `[object]` packet configuration parameters. see `config/examples/advanced/packetgen-*` for usage examples as there are just too many params to put them here. I'll only describe the general structure of the packet
- `packet.link` - `[layer]` tcp/ip level 1 (OSI level 2) configuration. currently only supports ethernet serialization but go runtime doesn't have a way to send custom ethernet frames so it's not advised to use it
- `packet.network` - `[layer]` tcp/ip level 2 (OSI level 3) configuration. supports `ipv4` and `ipv6` protocols. see `src/core/packetgen/network.go` for all the available options
- `packet.transport` - `[layer]` tcp/ip level 3 (OSI level 4) configuration. supports `tcp` and `udp` protocols. see `src/core/packetgen/transport.go` for all the available options
- `packet.payload` - `[layer]` the data that goes on top of other layers. for now it can be `raw` for custom crafted payload string (i.e. you can write an http request directly here), `dns`, and `icmpv4`, but last two are not fully tested yet

`dns-blast` args:

- `root_domain` - `[string]`
- `protocol` - `[string]` can be `udp`, `tcp`, or `tcp-tls`
- `seed_domains` - `[array]`

`slow-loris` - check `src/core/slowloris/slowloris.go` for reference

all the jobs have shared args:

- `interval_ms` - `[number]` interval between requests in milliseconds. Defaults to 0 (Care, in case of udp job it might generate the data faster than your OS/network card can process it)
- `count` - `[number]` limit the amount of requests to send with this job invocation. Defaults to 0 (no limit). Note: if config is refreshed before this limit is reached the job will be restarted and the counter will be reset

Almost every leaf `[string]` or `[object]` parameter can be templated with go template syntax. I've also added couple helper functions (list will be growing):

- `random_uuid`
- `random_int_n"`
- `random_int`
- `random_payload`
- `random_ip`
- `random_port`
- `random_mac_addr`
- `random_user_agent`
- `local_ip`
- `local_ipv4`
- `local_ipv6`
- `local_mac_addr`
- `resolve_host`
- `resolve_host_ipv4`
- `resolve_host_ipv6`
- `base64_encode`
- `base64_decode`
- `to_yaml`
- `from_yaml`
- `from_yaml_array`
- `to_json`
- `from_json`
- `from_json_array`
- `from_string_array`
- `join`
- `split`
- `get_url`
- `mod`
- `ctx_key`
- `split`
- `cookie_string`

Please refer to official go documentation and code in `src/utils/templates/` for these for now
