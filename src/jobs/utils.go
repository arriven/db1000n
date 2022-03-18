package jobs

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func logJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	var jobConfig struct {
		Text string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	logger.Info(templates.ParseAndExecute(logger, jobConfig.Text, ctx))

	return nil, nil
}

func setVarJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	var jobConfig struct {
		Value string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	return templates.ParseAndExecute(logger, jobConfig.Value, ctx), nil
}

func checkJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
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

func loopJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler(logger)

	var jobConfig struct {
		BasicJobConfig

		Job Config
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	for jobConfig.Next(ctx) {
		job := Get(jobConfig.Job.Type)
		if job == nil {
			return nil, fmt.Errorf("unknown job %q", jobConfig.Job.Type)
		}

		data, err := job(ctx, logger, globalConfig, jobConfig.Job.Args)
		if err != nil {
			return nil, fmt.Errorf("error running job: %w", err)
		}

		ctx = context.WithValue(ctx, templates.ContextKey("data."+jobConfig.Job.Name), data)
	}

	return nil, nil
}

func encryptedJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	if globalConfig.SkipEncrypted {
		return nil, fmt.Errorf("app is configured to skip encrypted jobs")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler(logger)

	var jobConfig struct {
		BasicJobConfig

		Format string
		Data   string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
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

	var jobCfg Config

	if err = utils.Unmarshal(decrypted, &jobCfg, jobConfig.Format); err != nil {
		return nil, err
	}

	job := Get(jobCfg.Type)
	if job == nil {
		return nil, fmt.Errorf("unknown job %q", jobCfg.Type)
	}

	return job(context.WithValue(ctx, templates.ContextKey(isEncryptedContextKey), true), zap.NewNop(), globalConfig, jobCfg.Args)
}
