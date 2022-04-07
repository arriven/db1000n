// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	"github.com/Arriven/db1000n/src/utils/metrics"
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
func InitRequest(c RequestConfig, req *fasthttp.Request) int64 {
	req.SetRequestURI(c.Path)
	req.Header.SetMethod(c.Method)
	req.SetBodyString(c.Body)
	// Add random user agent and configured headers
	req.Header.Set("user-agent", uarand.GetRandom())

	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	for key, value := range c.Cookies {
		req.Header.SetCookie(key, value)
	}

	dataSize, _ := req.WriteTo(metrics.NopWriter{})

	return dataSize
}

type Client interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
}

type StaticHostConfig struct {
	Addr  string
	IsTLS bool
}

// ClientConfig is a http client configuration structure
type ClientConfig struct {
	StaticHost      *StaticHostConfig
	TLSClientConfig *tls.Config
	Timeout         *time.Duration
	ReadTimeout     *time.Duration
	WriteTimeout    *time.Duration
	IdleTimeout     *time.Duration
	MaxIdleConns    *int
	ProxyURLs       string
}

// NewClient creates a fasthttp client based on the config.
func NewClient(ctx context.Context, clientConfig ClientConfig, logger *zap.Logger) Client {
	const (
		defaultMaxConnsPerHost = 1000
		defaultTimeout         = 90 * time.Second
	)

	timeout := utils.NonNilOrDefault(clientConfig.Timeout, defaultTimeout)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // This is intentional
	}
	if clientConfig.TLSClientConfig != nil {
		tlsConfig = clientConfig.TLSClientConfig
	}

	proxyFunc := utils.GetProxyFunc(templates.ParseAndExecute(logger, clientConfig.ProxyURLs, ctx), timeout, true)

	if clientConfig.StaticHost != nil {
		return &fasthttp.HostClient{
			Addr:                          clientConfig.StaticHost.Addr,
			IsTLS:                         clientConfig.StaticHost.IsTLS,
			MaxConnDuration:               timeout,
			ReadTimeout:                   utils.NonNilOrDefault(clientConfig.ReadTimeout, timeout),
			WriteTimeout:                  utils.NonNilOrDefault(clientConfig.WriteTimeout, timeout),
			MaxIdleConnDuration:           utils.NonNilOrDefault(clientConfig.IdleTimeout, timeout),
			MaxConns:                      utils.NonNilOrDefault(clientConfig.MaxIdleConns, defaultMaxConnsPerHost),
			NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
			DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
			DisablePathNormalizing:        true,
			TLSConfig:                     tlsConfig,
			Dial:                          dialViaProxyFunc(proxyFunc, "tcp"),
		}
	}

	return &fasthttp.Client{
		MaxConnDuration:               timeout,
		ReadTimeout:                   utils.NonNilOrDefault(clientConfig.ReadTimeout, timeout),
		WriteTimeout:                  utils.NonNilOrDefault(clientConfig.WriteTimeout, timeout),
		MaxIdleConnDuration:           utils.NonNilOrDefault(clientConfig.IdleTimeout, timeout),
		MaxConnsPerHost:               utils.NonNilOrDefault(clientConfig.MaxIdleConns, defaultMaxConnsPerHost),
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		TLSConfig:                     tlsConfig,
		Dial:                          dialViaProxyFunc(proxyFunc, "tcp"),
	}
}

func dialViaProxyFunc(proxyFunc utils.ProxyFunc, network string) fasthttp.DialFunc {
	// Return closure to select a random proxy on each call
	return func(addr string) (net.Conn, error) {
		return proxyFunc(network, addr)
	}
}
