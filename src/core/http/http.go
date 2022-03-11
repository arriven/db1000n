// Package http [allows sending customized http traffic]
package http

import (
	"crypto/tls"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/Arriven/db1000n/src/utils/templates"
	"github.com/corpix/uarand"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
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

// NewClient produces a client from a structure
func NewClient(clientConfig ClientConfig, debug bool) (client *fasthttp.Client) {
	timeout := 90 * time.Second
	if clientConfig.Timeout != nil {
		timeout = *clientConfig.Timeout
	}

	readTimeout := timeout
	if clientConfig.ReadTimeout != nil {
		readTimeout = *clientConfig.ReadTimeout
	}

	writeTimeout := timeout
	if clientConfig.WriteTimeout != nil {
		writeTimeout = *clientConfig.WriteTimeout
	}

	idleTimeout := timeout
	if clientConfig.IdleTimeout != nil {
		idleTimeout = *clientConfig.IdleTimeout
	}

	maxIdleConns := 1000
	if clientConfig.MaxIdleConns != nil {
		maxIdleConns = *clientConfig.MaxIdleConns
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if clientConfig.TLSClientConfig != nil {
		tlsConfig = clientConfig.TLSClientConfig
	}

	proxy := func() string { return "" }
	proxylist := templates.ParseAndExecute(clientConfig.ProxyURLs, nil)
	if proxylist != "" {
		if debug {
			log.Printf("List of proxies: %s", proxylist)
		}

		var proxyURLs = strings.Split(proxylist, ",")

		if debug {
			log.Printf("proxyURLs: %v", proxyURLs)
		}

		// Return random proxy from the list
		proxy = func() string {
			if len(proxyURLs) == 0 {
				return ""
			}

			return proxyURLs[rand.Intn(len(proxyURLs))]
		}
	}

	return &fasthttp.Client{
		ReadTimeout:                   readTimeout,
		WriteTimeout:                  writeTimeout,
		MaxConnDuration:               timeout,
		MaxIdleConnDuration:           idleTimeout,
		MaxConnsPerHost:               maxIdleConns,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		TLSConfig:                     tlsConfig,
		// increase DNS cache time to an hour instead of default minute
		Dial: fastHTTPProxyDial(proxy, fasthttpproxy.FasthttpProxyHTTPDialerTimeout(timeout)),
	}
}

func fastHTTPProxyDial(proxyFunc func() string, backup fasthttp.DialFunc) fasthttp.DialFunc {
	return func(addr string) (net.Conn, error) {
		proxy := proxyFunc()
		if proxy == "" {
			return backup(addr)
		}
		return fasthttpproxy.FasthttpSocksDialer(proxy)(addr)
	}
}
