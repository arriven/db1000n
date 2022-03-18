package jobs

import (
	"context"
	"fmt"
	"sync"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func sequenceJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler(logger)

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

		data, err := job(ctx, logger, globalConfig, cfg.Args)
		if err != nil {
			return nil, fmt.Errorf("error running job: %w", err)
		}

		ctx = context.WithValue(ctx, templates.ContextKey("data."+cfg.Name), data)
	}

	return nil, nil
}

func parallelJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler(logger)

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
			logger.Warn("Unknown job", zap.String("type", jobConfig.Jobs[i].Type))

			continue
		}

		if jobConfig.Jobs[i].Count < 1 {
			jobConfig.Jobs[i].Count = 1
		}

		for j := 0; j < jobConfig.Jobs[i].Count; j++ {
			wg.Add(1)

			go func(i int) {
				if _, err := job(ctx, logger, globalConfig, jobConfig.Jobs[i].Args); err != nil {
					logger.Error("error running job", zap.Error(err))
				}

				wg.Done()
			}(i)
		}
	}

	wg.Wait()

	return nil, nil
}
