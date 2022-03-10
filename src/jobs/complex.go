package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
	"github.com/mitchellh/mapstructure"
)

func sequenceJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig struct {
		BasicJobConfig

		Jobs []Config
	}
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %v", err)
	}

	for _, cfg := range jobConfig.Jobs {
		job, ok := Get(cfg.Type)
		if !ok {
			return nil, fmt.Errorf("unknown job %q", cfg.Type)
		}
		data, err := job(ctx, globalConfig, cfg.Args, debug)
		if err != nil {
			return nil, fmt.Errorf("error running job: %w", err)
		}
		ctx = context.WithValue(ctx, templates.ContextKey(cfg.Name), data)
	}
	return nil, nil
}

func parallelJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig struct {
		BasicJobConfig

		Jobs []Config
	}
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %v", err)
	}

	var wg sync.WaitGroup
	for i := range jobConfig.Jobs {
		job, ok := Get(jobConfig.Jobs[i].Type)
		if !ok {
			log.Printf("Unknown job %q", jobConfig.Jobs[i].Type)

			continue
		}

		if jobConfig.Jobs[i].Count < 1 {
			jobConfig.Jobs[i].Count = 1
		}

		for j := 0; j < jobConfig.Jobs[i].Count; j++ {
			wg.Add(1)

			go func(i int) {
				_, err := job(ctx, globalConfig, jobConfig.Jobs[i].Args, debug)
				if err != nil {
					log.Println("error running job:", err)
				}
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
	return nil, nil
}

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
