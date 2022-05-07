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

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils/metrics"
)

// "tcp" in config
func tcpJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	return rawnetJob(ctx, "tcp", args, globalConfig, a, logger)
}

// "udp" in config
func udpJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	return rawnetJob(ctx, "udp", args, globalConfig, a, logger)
}

func rawnetJob(ctx context.Context, protocol string, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (
	data any, err error,
) {
	packetgenArgs, err := parseRawNetJobArgs(logger, globalConfig, args, protocol)
	if err != nil {
		return nil, err
	}

	return packetgenJob(ctx, packetgenArgs, globalConfig, a, logger)
}

func parseRawNetJobArgs(logger *zap.Logger, globalConfig *GlobalConfig, args config.Args, protocol string) (
	result map[string]any, err error,
) {
	var jobConfig struct {
		BasicJobConfig

		Address   string
		Body      string
		ProxyURLs string
		Timeout   *time.Duration
	}

	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error decoding rawnet job config: %w", err)
	}

	packetgenArgs := make(map[string]any)
	for k, v := range args {
		packetgenArgs[k] = v
	}

	packetgenArgs["connection"] = map[string]any{
		"type": "net",
		"args": map[string]any{
			"protocol":   protocol,
			"address":    jobConfig.Address,
			"timeout":    jobConfig.Timeout,
			"proxy_urls": jobConfig.ProxyURLs,
		},
	}
	packetgenArgs["packet"] = map[string]any{
		"payload": map[string]any{
			"type": "raw",
			"data": map[string]any{
				"payload": jobConfig.Body,
			},
		},
	}

	return packetgenArgs, nil
}
