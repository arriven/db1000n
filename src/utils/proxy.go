package utils

import (
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

	u, err := url.Parse(proxies[rand.Intn(len(proxies))]) //nolint:gosec // Cryptographically secure random not required
	if err != nil {
		return proxy.FromEnvironmentUsing(direct).Dial
	}

	client, err := proxy.FromURL(u, direct)
	if err != nil {
		return proxy.FromEnvironmentUsing(direct).Dial
	}

	return client.Dial
}
