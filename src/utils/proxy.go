package utils

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type ProxyFunc func(network, addr string) (net.Conn, error)

func GetProxyFunc(proxyURLs string, timeout time.Duration) ProxyFunc {
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

		client, err := proxy.FromURL(u, direct)
		if err != nil {
			return nil, fmt.Errorf("error building proxy %v: %w", u.String(), err)
		}

		return client.Dial(network, addr)
	}
}
