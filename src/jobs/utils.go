package jobs

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/mitchellh/mapstructure"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func logJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	var jobConfig struct {
		Text string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	log.Println(templates.ParseAndExecute(jobConfig.Text, ctx))

	return nil, nil
}

func setVarJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	var jobConfig struct {
		Value string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	return templates.ParseAndExecute(jobConfig.Value, ctx), nil
}

func checkJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	var jobConfig struct {
		Value string
	}

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	if templates.ParseAndExecute(jobConfig.Value, ctx) != "true" {
		return nil, fmt.Errorf("validation failed %v", jobConfig.Value)
	}

	return nil, nil
}

func loopJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

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

		data, err := job(ctx, globalConfig, jobConfig.Job.Args, debug)
		if err != nil {
			return nil, fmt.Errorf("error running job: %w", err)
		}

		ctx = context.WithValue(ctx, templates.ContextKey("data."+jobConfig.Job.Name), data)
	}

	return nil, nil
}

func encryptedJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	if globalConfig.SkipEncrypted {
		return nil, fmt.Errorf("app is configured to skip encrypted jobs")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

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

	return job(ctx, globalConfig, jobCfg.Args, debug)
}
