package jobs

import (
	"context"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/packetgen"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func packetgenJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler(logger)

	type packetgenJobConfig struct {
		BasicJobConfig
		Packet     map[string]interface{}
		Connection packetgen.ConnectionConfig
	}

	var jobConfig packetgenJobConfig

	if err := utils.Decode(args, &jobConfig); err != nil {
		logger.Debug("error parsing job config", zap.Error(err))

		return nil, err
	}

	rawConn, err := packetgen.OpenRawConnection(jobConfig.Connection)
	if err != nil {
		logger.Debug("error building raw connection", zap.Error(err))

		return nil, err
	}

	packetTpl, err := templates.ParseMapStruct(jobConfig.Packet)
	if err != nil {
		logger.Debug("error parsing packet", zap.Error(err))

		return nil, err
	}

	payloadBuf := gopacket.NewSerializeBuffer()

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {
		if err := payloadBuf.Clear(); err != nil {
			logger.Debug("error clearing payload buffer", zap.Error(err))

			return nil, err
		}

		packetConfigRaw := packetTpl.Execute(logger, ctx)
		logger.Debug("rendered packet config template", zap.Reflect("config", packetConfigRaw))

		var packetConfig packetgen.PacketConfig
		if err := utils.Decode(packetConfigRaw, &packetConfig); err != nil {
			logger.Debug("error parsing packet config", zap.Error(err))

			return nil, err
		}

		packet, err := packetConfig.Build()
		if err != nil {
			logger.Debug("error building packet", zap.Error(err))

			return nil, err
		}

		if err = packet.Serialize(payloadBuf); err != nil {
			logger.Debug("error serializing packet", zap.Error(err))

			return nil, err
		}

		if _, err = rawConn.WriteTo(payloadBuf.Bytes(), nil, &net.IPAddr{IP: packet.IP()}); err != nil {
			logger.Debug("error sending packet", zap.Error(err))

			return nil, err
		}

		trafficMonitor.Add(uint64(len(payloadBuf.Bytes())))
	}

	return nil, nil
}
