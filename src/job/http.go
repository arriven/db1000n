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

package job

import (
	"context"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/http"
	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

type httpJobConfig struct {
	BasicJobConfig

	Request map[string]any
	Client  map[string]any // See HTTPClientConfig
}

func singleRequestJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	_, clientConfig, requestTpl, err := getHTTPJobConfigs(ctx, args, *globalConfig, logger)
	if err != nil {
		return nil, err
	}

	var requestConfig http.RequestConfig
	if err := utils.Decode(requestTpl.Execute(logger, ctx), &requestConfig); err != nil {
		return nil, err
	}

	client := http.NewClient(ctx, *clientConfig, logger)

	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	logger.Info("single http request", zap.String("target", requestConfig.Path))

	http.InitRequest(requestConfig, req)

	if err = sendFastHTTPRequest(client, req, resp); err != nil {
		if a != nil {
			a.Inc(target(req.URI()), metrics.RequestsAttemptedStat).Flush()
		}

		return nil, err
	}

	requestSize, _ := req.WriteTo(metrics.NopWriter{})

	if a != nil {
		a.AddStats(target(req.URI()), metrics.NewStats(1, 1, 1, uint64(requestSize))).Flush()
	}

	headers, cookies := make(map[string]string), make(map[string]string)

	resp.Header.VisitAll(headerLoaderFunc(headers))
	resp.Header.VisitAllCookie(cookieLoaderFunc(cookies, logger))

	return map[string]any{
		"response": map[string]any{
			"body":        string(resp.Body()),
			"status_code": resp.StatusCode(),
			"headers":     headers,
			"cookies":     cookies,
		},
		"error": err,
	}, nil
}

func headerLoaderFunc(headers map[string]string) func(key []byte, value []byte) {
	return func(key []byte, value []byte) {
		headers[string(key)] = string(value)
	}
}

func cookieLoaderFunc(cookies map[string]string, logger *zap.Logger) func(key []byte, value []byte) {
	return func(key []byte, value []byte) {
		c := fasthttp.AcquireCookie()
		defer fasthttp.ReleaseCookie(c)

		if err := c.ParseBytes(value); err != nil {
			return
		}

		if expire := c.Expire(); expire != fasthttp.CookieExpireUnlimited && expire.Before(time.Now()) {
			logger.Debug("cookie from the request expired", zap.ByteString("cookie", key))

			return
		}

		cookies[string(key)] = string(c.Value())
	}
}

func fastHTTPJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobConfig, clientConfig, requestTpl, err := getHTTPJobConfigs(ctx, args, *globalConfig, logger)
	if err != nil {
		return nil, err
	}

	backoffController := utils.BackoffController{BackoffConfig: utils.NonNilOrDefault(jobConfig.Backoff, globalConfig.Backoff)}
	client := http.NewClient(ctx, *clientConfig, logger)

	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
	}()

	resp.SkipBody = true

	logger.Info("attacking", zap.Any("target", jobConfig.Request["path"]))

	for jobConfig.Next(ctx) {
		var requestConfig http.RequestConfig
		if err := utils.Decode(requestTpl.Execute(logger, ctx), &requestConfig); err != nil {
			return nil, fmt.Errorf("error executing request template: %w", err)
		}

		http.InitRequest(requestConfig, req)

		if err := sendFastHTTPRequest(client, req, resp); err != nil {
			logger.Debug("error sending request", zap.Error(err), zap.Any("args", args))

			if a != nil {
				a.Inc(target(req.URI()), metrics.RequestsAttemptedStat).Flush()
			}

			utils.Sleep(ctx, backoffController.Increment().GetTimeout())

			continue
		}

		requestSize, _ := req.WriteTo(metrics.NopWriter{})

		if a != nil {
			a.AddStats(target(req.URI()), metrics.NewStats(1, 1, 1, uint64(requestSize))).Flush()
		}

		backoffController.Reset()
	}

	return nil, nil
}

func target(uri *fasthttp.URI) string { return string(uri.Scheme()) + "://" + string(uri.Host()) }

func getHTTPJobConfigs(ctx context.Context, args config.Args, global GlobalConfig, logger *zap.Logger) (
	cfg *httpJobConfig, clientCfg *http.ClientConfig, requestTpl *templates.MapStruct, err error,
) {
	var jobConfig httpJobConfig
	if err := ParseConfig(&jobConfig, args, global); err != nil {
		return nil, nil, nil, fmt.Errorf("error parsing job config: %w", err)
	}

	var clientConfig http.ClientConfig
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(logger, jobConfig.Client, ctx), &clientConfig); err != nil {
		return nil, nil, nil, fmt.Errorf("error parsing client config: %w", err)
	}

	if global.ProxyURLs != "" {
		clientConfig.ProxyURLs = templates.ParseAndExecute(logger, global.ProxyURLs, ctx)
	}

	requestTpl, err = templates.ParseMapStruct(jobConfig.Request)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error parsing request config: %w", err)
	}

	return &jobConfig, &clientConfig, requestTpl, nil
}

func sendFastHTTPRequest(client http.Client, req *fasthttp.Request, resp *fasthttp.Response) error {
	if err := client.Do(req, resp); err != nil {
		metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusFail)

		return err
	}

	metrics.IncHTTP(string(req.Host()), string(req.Header.Method()), metrics.StatusSuccess)

	return nil
}
