package jobs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Arriven/db1000n/src/slowloris"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/mitchellh/mapstructure"
)

func slowLorisJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) error {
	defer utils.PanicHandler()

	var jobConfig *slowloris.Config
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return err
	}

	if len(jobConfig.Path) == 0 {
		return errors.New("path is empty")
	}

	if jobConfig.ContentLength == 0 {
		jobConfig.ContentLength = 1000 * 1000
	}

	if jobConfig.DialWorkersCount == 0 {
		jobConfig.DialWorkersCount = 10
	}

	if jobConfig.RampUpInterval == 0 {
		jobConfig.RampUpInterval = 1 * time.Second
	}

	if jobConfig.SleepInterval == 0 {
		jobConfig.SleepInterval = 10 * time.Second
	}

	if jobConfig.Duration == 0 {
		jobConfig.Duration = 10 * time.Second
	}

	shouldStop := make(chan bool)
	go func() {
		<-ctx.Done()
		close(shouldStop)
	}()

	if debug {
		log.Printf("sending slow loris with params: %v", jobConfig)
	}

	return slowloris.Start(shouldStop, jobConfig)
}
