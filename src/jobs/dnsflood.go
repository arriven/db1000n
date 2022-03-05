package jobs

import (
	"context"
	"encoding/json"
	"github.com/Arriven/db1000n/dnsflood"
	"github.com/Arriven/db1000n/src/logs"
	"github.com/Arriven/db1000n/src/utils"
)

func dnsFloodJob(ctx context.Context, l *logs.Logger, args Args) error {
	defer utils.PanicHandler()
	var jobConfig *dnsflood.Config
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}

	shouldStop := make(chan bool)
	go func() {
		<-ctx.Done()
		shouldStop <- true
	}()
	l.Debug("sending DNS flood with params: %v", jobConfig)

	return dnsflood.Start(shouldStop, l, jobConfig)
}
