# DNS Blast

An adoption of DNS Stress tool to validate the DNS server throughput.
Original source codes: [https://github.com/sandeeprenjith/dnsblast](https://github.com/sandeeprenjith/dnsblast).

Heavily relies on the idea of [Distinct Heavy Hitter attack](https://faculty.idc.ac.il/bremler/Papers/HotWeb_18.pdf):

Sends DNS queries with QType set to 'A'.

## Job parameters

- `root_domain` - Root domain to use which will be queried for nameservers
- `protocol` - DNS net. protocol (`udp` as default, [`udp`, `tcp`, `tcp-tls`] supported)
- `seed_domains` - Domain names to use as base of DNS query (no default, at least one required, like `yahoo.com`)
- `parallel_queries` - Number of DNS queries to send between delays
- `interval_ms` - (inherited) delay in MS between query loop iteration

## A note from the Author

_Use only to test your own services, do not abuse services that you do not own!_

Research references:

- [https://www.ietf.org/proceedings/91/slides/slides-91-dprive-5.pdf](https://www.ietf.org/proceedings/91/slides/slides-91-dprive-5.pdf)
- [https://github.com/refraction-networking/utls](https://github.com/refraction-networking/utls)
- [https://ja3er.com/about.html](https://ja3er.com/about.html)
- [https://habr.com/ru/post/596411/](https://habr.com/ru/post/596411/)
