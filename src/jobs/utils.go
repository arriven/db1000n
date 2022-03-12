package jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
	"github.com/mitchellh/mapstructure"
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
