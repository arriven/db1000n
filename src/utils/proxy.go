package utils

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp/fasthttpproxy"
	"golang.org/x/net/proxy"
	"h12.io/socks"
)

type ProxyFunc func(network, addr string) (net.Conn, error)

type ProxyParams struct {
	URLs      string
	LocalAddr string
	Interface string
	Timeout   time.Duration
}

// this won't work for udp payloads but if people use proxies they might not want to have their ip exposed
// so it's probably better to fail instead of routing the traffic directly
func GetProxyFunc(params ProxyParams, protocol string) ProxyFunc {
	direct := &net.Dialer{Timeout: params.Timeout, LocalAddr: resolveAddr(protocol, params.LocalAddr), Control: BindToInterface(params.Interface)}
	if params.URLs == "" {
		return proxy.FromEnvironmentUsing(direct).Dial
	}

	proxies := strings.Split(params.URLs, ",")

	// We need to dial new proxy on each call
	return func(network, addr string) (net.Conn, error) {
		u, err := url.Parse(proxies[rand.Intn(len(proxies))]) //nolint:gosec // Cryptographically secure random not required
		if err != nil {
			return nil, fmt.Errorf("error building proxy %v: %w", u.String(), err)
		}

		switch u.Scheme {
		case "socks5", "socks5h":
			client, err := proxy.FromURL(u, direct)
			if err != nil {
				return nil, fmt.Errorf("error building proxy %v: %w", u.String(), err)
			}

			return client.Dial(network, addr)
		case "socks4", "socks4a":
			return socks.Dial(u.String())(network, addr)
		default:
			// Not all http proxies support tunneling so it's safer to skip them for raw tcp payload
			if protocol == "http" {
				return fasthttpproxy.FasthttpHTTPDialerTimeout(u.Host, params.Timeout)(addr)
			}

			return nil, fmt.Errorf("unsupported proxy scheme %v", u.Scheme)
		}
	}
}

func resolveAddr(protocol, addr string) net.Addr {
	if addr == "" {
		return nil
	}

	var zone string

	// handle ipv6 zone
	if strings.Contains(addr, "%") {
		split := strings.Split(addr, "%")
		addr, zone = split[0], split[1]
	}

	ip := net.ParseIP(addr)

	switch protocol {
	case "tcp", "tcp4", "tcp6", "http":
		return &net.TCPAddr{IP: ip, Zone: zone}
	case "udp", "udp4", "udp6":
		return &net.UDPAddr{IP: ip, Zone: zone}
	default:
		return &net.IPAddr{IP: ip}
	}
}
