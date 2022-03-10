package jobs

import (
	"context"
	"fmt"
	"log"

	"github.com/Arriven/db1000n/src/utils/templates"
	"github.com/mitchellh/mapstructure"
)

func logJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	var jobConfig struct {
		Text string
	}
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %v", err)
	}
	log.Println(templates.ParseAndExecute(jobConfig.Text, ctx))
	return nil, nil
}

func setVarJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	var jobConfig struct {
		Value string
	}
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %v", err)
	}
	return templates.ParseAndExecute(jobConfig.Value, ctx), nil
}

func checkJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	var jobConfig struct {
		Value string
	}
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %v", err)
	}
	if templates.ParseAndExecute(jobConfig.Value, ctx) != "true" {
		return nil, fmt.Errorf("validation failed %v", jobConfig.Value)
	}
	return nil, nil
}
