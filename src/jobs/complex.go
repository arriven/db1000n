package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/mitchellh/mapstructure"
)

func sequenceJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig struct {
		BasicJobConfig

		Jobs []Config
	}
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return fmt.Errorf("error parsing job config: %v", err)
	}

	for _, cfg := range jobConfig.Jobs {
		job, ok := Get(cfg.Type)
		if !ok {
			return fmt.Errorf("unknown job %q", cfg.Type)
		}
		err := job(ctx, globalConfig, cfg.Args, debug)
		if err != nil {
			return fmt.Errorf("error running job: %w", err)
		}
	}
	return nil
}

func parallelJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	var jobConfig struct {
		BasicJobConfig

		Jobs []Config
	}
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return fmt.Errorf("error parsing job config: %v", err)
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
				err := job(ctx, globalConfig, jobConfig.Jobs[i].Args, debug)
				if err != nil {
					log.Println("error running job:", err)
				}
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
	return nil
}
