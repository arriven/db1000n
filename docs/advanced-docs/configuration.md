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
- `request.cookies` - `[object]` key-value map of http cookies
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

Warning: `packetgen` requires root privileges to run

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
- `tcp.flags` - `[object]` flags for tcp (every flag has its respective name)

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
- `local_mac_addr`
- `resolve_host`
- `base64_encode`
- `base64_decode`
- `json_encode`
- `json_decode`
- `yaml_encode`
- `yaml_decode`
- `join`
- `get_url`
- `proxylist_url`
- `get_proxylist`
- `get_proxylist_by_url`
- `mod`
- `ctx_key`
- `split`
- `cookie_string`

Please refer to official go documentation and code for these for now
