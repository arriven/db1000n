package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mitchellh/mapstructure"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
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
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	for _, cfg := range jobConfig.Jobs {
		job := Get(cfg.Type)
		if job == nil {
			return nil, fmt.Errorf("unknown job %q", cfg.Type)
		}

		data, err := job(ctx, globalConfig, cfg.Args, debug)
		if err != nil {
			return nil, fmt.Errorf("error running job: %w", err)
		}

		ctx = context.WithValue(ctx, templates.ContextKey("data."+cfg.Name), data)
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
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	var wg sync.WaitGroup
	for i := range jobConfig.Jobs {
		job := Get(jobConfig.Jobs[i].Type)
		if job == nil {
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
