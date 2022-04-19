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
	"encoding/base64"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/robertkrimen/otto"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func logJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (
	data any, err error, //nolint:unparam // data is here to match Job
) {
	var jobConfig struct {
		Text string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	logger.Info(templates.ParseAndExecute(logger, jobConfig.Text, ctx))

	return nil, nil
}

func setVarJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	var jobConfig struct {
		Value string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	return templates.ParseAndExecute(logger, jobConfig.Value, ctx), nil
}

func checkJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (
	data any, err error, //nolint:unparam // data is here to match Job
) {
	var jobConfig struct {
		Value string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	if templates.ParseAndExecute(logger, jobConfig.Value, ctx) != "true" {
		return nil, fmt.Errorf("validation failed %v", jobConfig.Value)
	}

	return nil, nil
}

func sleepJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	var jobConfig struct {
		Value time.Duration
	}

	if err := utils.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	time.Sleep(jobConfig.Value)

	return nil, nil
}

func discardErrorJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (
	data any, err error, //nolint:unparam // data is here to match Job
) {
	var jobConfig struct {
		BasicJobConfig

		Job config.Config
	}

	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	job := Get(jobConfig.Job.Type)
	if job == nil {
		logger.Debug("unknown job, discarding", zap.String("job", jobConfig.Job.Type))

		return nil, nil
	}

	data, err = job(ctx, jobConfig.Job.Args, globalConfig, a, logger)
	if err != nil {
		logger.Debug("discarded error", zap.Error(err))
	}

	return data, nil
}

func timeoutJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (
	data any, err error, //nolint:unparam // data is here to match Job
) {
	var jobConfig struct {
		BasicJobConfig

		Timeout time.Duration
		Job     config.Config
	}

	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, jobConfig.Timeout)
	defer cancel()

	job := Get(jobConfig.Job.Type)

	if job == nil {
		return nil, fmt.Errorf("unknown job %q", jobConfig.Job.Type)
	}

	return job(ctx, jobConfig.Job.Args, globalConfig, a, logger)
}

func loopJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var jobConfig struct {
		BasicJobConfig

		Job config.Config
	}

	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	for jobConfig.Next(ctx) {
		job := Get(jobConfig.Job.Type)
		if job == nil {
			return nil, fmt.Errorf("unknown job %q", jobConfig.Job.Type)
		}

		data, err := job(ctx, jobConfig.Job.Args, globalConfig, a, logger)
		if err != nil {
			return nil, fmt.Errorf("error running job: %w", err)
		}

		ctx = context.WithValue(ctx, templates.ContextKey("data."+jobConfig.Job.Name), data)
	}

	return nil, nil
}

func jsJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (
	data any, err error,
) {
	var jobConfig struct {
		Script string
		Data   map[string]any
	}

	if err := mapstructure.Decode(templates.ParseAndExecuteMapStruct(logger, args, ctx), &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	vm := otto.New()

	for key, value := range jobConfig.Data {
		if err := vm.Set(key, value); err != nil {
			return nil, fmt.Errorf("error setting script data: %w", err)
		}
	}

	return vm.Run(jobConfig.Script)
}

const isEncryptedContextKey = "is_in_encrypted_context"

func EncryptedContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, templates.ContextKey(isEncryptedContextKey), true)
}

func isInEncryptedContext(ctx context.Context) bool {
	return ctx.Value(templates.ContextKey(isEncryptedContextKey)) != nil
}

func encryptedJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	if globalConfig.SkipEncrypted {
		return nil, fmt.Errorf("app is configured to skip encrypted jobs")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var jobConfig struct {
		BasicJobConfig

		Format string
		Data   string
	}

	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(jobConfig.Data)
	if err != nil {
		return nil, err
	}

	decrypted, err := utils.Decrypt(decoded)
	if err != nil {
		return nil, err
	}

	var jobCfg config.Config

	if err = utils.Unmarshal(decrypted, &jobCfg, jobConfig.Format); err != nil {
		return nil, err
	}

	job := Get(jobCfg.Type)
	if job == nil {
		return nil, fmt.Errorf("unknown job %q", jobCfg.Type)
	}

	return job(EncryptedContext(ctx), jobCfg.Args, globalConfig, a, zap.NewNop())
}
