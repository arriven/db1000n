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

func singleRequestJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig httpJobConfig
	if err := utils.Decode(args, &jobConfig); err != nil {
		log.Printf("Error parsing job config: %v", err)
		return nil, err
	}
	var clientConfig http.ClientConfig
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(jobConfig.Client, ctx), &clientConfig); err != nil {
		log.Printf("Error parsing client config: %v", err)
		return nil, err
	}
	if globalConfig.ProxyURL != "" {
		clientConfig.ProxyURLs = globalConfig.ProxyURL
	}
	client := http.NewClient(clientConfig, debug)

	var requestConfig http.RequestConfig
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(jobConfig.Request, ctx), &requestConfig); err != nil {
		log.Printf("Error parsing request config: %v", err)
		return nil, err
	}
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	log.Printf("Sent single http request to %v", requestConfig.Path)
	dataSize := http.InitRequest(requestConfig, req)

	metrics.Default.Write(metrics.Traffic, uuid.New().String(), uint64(dataSize))
	err = sendFastHTTPRequest(client, req, resp, debug)
	if err == nil {
		metrics.Default.Write(metrics.ProcessedTraffic, uuid.New().String(), uint64(dataSize))
	}
	headers := make(map[string]string)
	resp.Header.VisitAll(func(key []byte, value []byte) {
		headers[string(key)] = string(value)
	})
	cookies := make(map[string]string)
	resp.Header.VisitAllCookie(func(key []byte, value []byte) {
		c := fasthttp.AcquireCookie()
		defer fasthttp.ReleaseCookie(c)

		err := c.ParseBytes(value)
		if err != nil {
			return
		}

		if expire := c.Expire(); expire != fasthttp.CookieExpireUnlimited && expire.Before(time.Now()) {
			if debug {
				log.Println("cookie from request expired:", string(key))
			}
			return
		}
		cookies[string(key)] = string(c.Value())
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
	var clientConfig http.ClientConfig
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(jobConfig.Client, ctx), &clientConfig); err != nil {
		log.Printf("Error parsing client config: %v", err)
		return nil, err
	}
	if globalConfig.ProxyURL != "" {
		clientConfig.ProxyURLs = globalConfig.ProxyURL
	}
	client := http.NewClient(clientConfig, debug)

	requestTpl, err := templates.ParseMapStruct(jobConfig.Request)
	if err != nil {
		log.Printf("Error parsing request config: %v", err)
		return nil, err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)
	processedTrafficMonitor := metrics.Default.NewWriter(metrics.ProcessedTraffic, uuid.NewString())
	go processedTrafficMonitor.Update(ctx, time.Second)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	for jobConfig.Next(ctx) {
		var requestConfig http.RequestConfig
		if err := utils.Decode(requestTpl.Execute(ctx), &requestConfig); err != nil {
			log.Printf("Error executing request template: %v", err)
			return nil, err
		}
		log.Printf("Sent single http request to %v", requestConfig.Path)
		dataSize := http.InitRequest(requestConfig, req)

		trafficMonitor.Add(uint64(dataSize))
		if err := sendFastHTTPRequest(client, req, nil, debug); err != nil {
			if debug {
				log.Printf("Error sending request %v: %v", req, err)
			}
		} else {
			processedTrafficMonitor.Add(uint64(dataSize))
		}
	}

	return nil, nil
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
