// Package http [allows sending customized http traffic]
package http

import (
	"crypto/tls"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/corpix/uarand"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils/templates"
)

// RequestConfig is a struct representing the config of a single request
type RequestConfig struct {
	Path    string
	Method  string
	Body    string
	Headers map[string]string
	Cookies map[string]string
}

// InitRequest is used to populate data from request config to fasthttp.Request
func InitRequest(c RequestConfig, req *fasthttp.Request) int {
	dataSize := len(c.Method) + len(c.Path) + len(c.Body) // Rough uploaded data size for reporting

	req.SetRequestURI(c.Path)
	req.Header.SetMethod(c.Method)
	req.SetBodyString(c.Body)
	// Add random user agent and configured headers
	req.Header.Set("user-agent", uarand.GetRandom())

	for key, value := range c.Headers {
		req.Header.Set(key, value)
		dataSize += len(key) + len(value)
	}

	for key, value := range c.Cookies {
		req.Header.SetCookie(key, value)
		dataSize += len(key) + len(value)
	}

	return dataSize
}

// ClientConfig is a http client configuration structure
type ClientConfig struct {
	TLSClientConfig *tls.Config    `mapstructure:"tls_config,omitempty"`
	Timeout         *time.Duration `mapstructure:"timeout"`
	ReadTimeout     *time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    *time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     *time.Duration `mapstructure:"idle_timeout"`
	MaxIdleConns    *int           `mapstructure:"max_idle_connections"`
	ProxyURLs       string         `mapstructure:"proxy_urls"`
}

// NewClient creates a fasthttp client based on the config.
func NewClient(clientConfig ClientConfig, logger *zap.Logger) *fasthttp.Client {
	const (
		defaultMaxConnsPerHost = 1000
		defaultTimeout         = 90 * time.Second
	)

	timeout := nonNilDurationOrDefault(clientConfig.Timeout, defaultTimeout)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // This is intentional
	}
	if clientConfig.TLSClientConfig != nil {
		tlsConfig = clientConfig.TLSClientConfig
	}

	return &fasthttp.Client{
		MaxConnDuration:               timeout,
		ReadTimeout:                   nonNilDurationOrDefault(clientConfig.ReadTimeout, timeout),
		WriteTimeout:                  nonNilDurationOrDefault(clientConfig.WriteTimeout, timeout),
		MaxIdleConnDuration:           nonNilDurationOrDefault(clientConfig.IdleTimeout, timeout),
		MaxConnsPerHost:               nonNilIntOrDefault(clientConfig.MaxIdleConns, defaultMaxConnsPerHost),
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		TLSConfig:                     tlsConfig,
		Dial: dialViaProxyFunc(templates.ParseAndExecute(logger, clientConfig.ProxyURLs, nil),
			fasthttpproxy.FasthttpProxyHTTPDialerTimeout(timeout),
			logger),
	}
}

func dialViaProxyFunc(proxyListCSV string, backup fasthttp.DialFunc, logger *zap.Logger) fasthttp.DialFunc {
	if len(proxyListCSV) == 0 {
		return backup
	}

	proxyURLs := strings.Split(proxyListCSV, ",")

	logger.Debug("proxyURLs parsed", zap.Strings("proxyURLs", proxyURLs))

	// Return closure to select a random proxy on each call
	return func(addr string) (net.Conn, error) {
		return fasthttpproxy.FasthttpSocksDialer(proxyURLs[rand.Intn(len(proxyURLs))])(addr) //nolint:gosec // Cryptographically secure random not required
	}
}

func nonNilDurationOrDefault(d *time.Duration, dflt time.Duration) time.Duration {
	if d != nil {
		return *d
	}

	return dflt
}

func nonNilIntOrDefault(i *int, dflt int) int {
	if i != nil {
		return *i
	}

	return dflt
}
