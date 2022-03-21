// Package http [allows sending customized http traffic]
package http

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/corpix/uarand"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
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
func NewClient(ctx context.Context, clientConfig ClientConfig, logger *zap.Logger) *fasthttp.Client {
	const (
		defaultMaxConnsPerHost = 1000
		defaultTimeout         = 90 * time.Second
	)

	timeout := utils.NonNilDurationOrDefault(clientConfig.Timeout, defaultTimeout)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // This is intentional
	}
	if clientConfig.TLSClientConfig != nil {
		tlsConfig = clientConfig.TLSClientConfig
	}

	return &fasthttp.Client{
		MaxConnDuration:               timeout,
		ReadTimeout:                   utils.NonNilDurationOrDefault(clientConfig.ReadTimeout, timeout),
		WriteTimeout:                  utils.NonNilDurationOrDefault(clientConfig.WriteTimeout, timeout),
		MaxIdleConnDuration:           utils.NonNilDurationOrDefault(clientConfig.IdleTimeout, timeout),
		MaxConnsPerHost:               utils.NonNilIntOrDefault(clientConfig.MaxIdleConns, defaultMaxConnsPerHost),
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		TLSConfig:                     tlsConfig,
		Dial:                          dialViaProxyFunc(utils.GetProxyFunc(templates.ParseAndExecute(logger, clientConfig.ProxyURLs, ctx), timeout), "tcp"),
	}
}

func dialViaProxyFunc(proxyFunc utils.ProxyFunc, network string) fasthttp.DialFunc {
	// Return closure to select a random proxy on each call
	return func(addr string) (net.Conn, error) {
		return proxyFunc(network, addr)
	}
}
