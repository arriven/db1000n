# DNS Blast

An adoption of DNS Stress tool to validate the DNS server throughput.
Original source codes: [https://github.com/sandeeprenjith/dnsblast](https://github.com/sandeeprenjith/dnsblast).

Heavily relies on the idea of [Distinct Heavy Hitter attack](https://faculty.idc.ac.il/bremler/Papers/HotWeb_18.pdf):

Sends DNS queries with QType set to 'A'.

## Job parameters

* `target_server_ip` - Target server IP address (no default value)
* `target_server_port` - Target server port (`53` for UDP & TCP, `853` for TCP-TLS)
* `protocol` - DNS net. protocol (`udp` as default, [`udp`, `tcp`, `tcp-tls`] supported)
* `seed_domains` - Domain names to use as base of DNS query (no default, at least one required, like `yahoo.com`)
* `parallel_queries` - Number of DNS queries to send between delays
* `interval_ms` - (inherited) delay in MS between query loop iteration 
