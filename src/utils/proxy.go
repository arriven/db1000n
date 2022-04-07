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

func GetProxyFunc(proxyURLs string, timeout time.Duration, httpEnabled bool) ProxyFunc {
	direct := &net.Dialer{Timeout: timeout}
	if proxyURLs == "" {
		return proxy.FromEnvironmentUsing(direct).Dial
	}

	proxies := strings.Split(proxyURLs, ",")

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
			if httpEnabled {
				return fasthttpproxy.FasthttpHTTPDialerTimeout(u.Host, timeout)(addr)
			}

			return nil, fmt.Errorf("unsupported proxy scheme %v", u.Scheme)
		}
	}
}
