package jobs

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/slowloris"
	"github.com/Arriven/db1000n/src/utils"
)

func slowLorisJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler(logger)

	var jobConfig *slowloris.Config
	if err := utils.Decode(args, &jobConfig); err != nil {
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

	jobConfig.ClientID = globalConfig.ClientID

	shouldStop := make(chan bool)

	go func() {
		<-ctx.Done()
		close(shouldStop)
	}()

	logger.Debug("sending flow loris with params", zap.Reflect("params", jobConfig))

	return nil, slowloris.Start(shouldStop, logger, jobConfig)
}
