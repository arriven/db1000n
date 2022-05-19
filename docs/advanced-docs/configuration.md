# Configuration

## Command Line reference

```text
Usage of db1000n:
  -b string
      raw backup config in case the primary one is unavailable
  -c string
      path to config files, separated by a comma, each path can be a web endpoint (default "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.v0.7.json")
  -country-list string
      comma-separated list of countries (default "Ukraine")
  -debug
      enable debug level logging
  -enable-primitive
      set to true if you want to run primitive jobs that are less resource-efficient (default true)
  -enable-self-update
      Enable the application automatic updates on the startup
  -format string
      config format (default "yaml")
  -h  print help message and exit
  -pprof string
      enable pprof
  -prometheus_gateways string
      Comma separated list of prometheus push gateways (default "https://178.62.78.144:9091,https://46.101.26.43:9091,https://178.62.33.149:9091")
  -prometheus_on
      Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>) (default true)
  -proxy string
      system proxy to set by default (can be a comma-separated list or a template)
  -refresh-interval duration
      refresh timeout for updating the config (default 1m0s)
  -restart-on-update
      Allows application to restart upon successful update (ignored if auto-update is disabled) (default true)
  -scale int
      used to scale the amount of jobs being launched, effect is similar to launching multiple instances at once (default 1)
  -self-update-check-frequency duration
      How often to run auto-update checks (default 24h0m0s)
  -skip-encrypted
      set to true if you want to only run plaintext jobs from the config for security considerations
  -skip-update-check-on-start
      Allows to skip the update check at the startup (usually set automatically by the previous version)
  -strict-country-check
      enable strict country check; will also exit if IP can't be determined
  -updater-destination-config string
      Destination config file to write (only applies if updater-mode is enabled (default "config/config.json")
  -updater-mode
      Only run config updater
```

Almost all of these parameters can also be set via environment variables

## Config file reference

This doc gets outdated frequently as the project is under active development but you can always check up to date configuration examples in `examples/config` folder

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
- `request.cookies` - `[object]` key-value map of http cookies (you can still set cookies directly via the header with `cookie_string` template function or statically, see `examples/config/advanced/ddos-guard.yaml` for an example)
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
- `packet` - `[object]` packet configuration parameters. see `examples/config/advanced/packetgen-*` for usage examples as there are just too many params to put them here. I'll only describe the general structure of the packet
- `packet.link` - `[layer]` tcp/ip level 1 (OSI level 2) configuration. currently only supports ethernet serialization but go runtime doesn't have a way to send custom ethernet frames so it's not advised to use it
- `packet.network` - `[layer]` tcp/ip level 2 (OSI level 3) configuration. supports `ipv4` and `ipv6` protocols. see `src/core/packetgen/network.go` for all the available options
- `packet.transport` - `[layer]` tcp/ip level 3 (OSI level 4) configuration. supports `tcp` and `udp` protocols. see `src/core/packetgen/transport.go` for all the available options
- `packet.payload` - `[layer]` the data that goes on top of other layers. for now it can be `raw` for custom crafted payload string (i.e. you can write an http request directly here), `dns`, and `icmpv4`, but last two are not fully tested yet

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
