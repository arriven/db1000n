// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package job

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/packetgen"
	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

type packetgenJobConfig struct {
	BasicJobConfig
	Packet     *templates.MapStruct
	Connection packetgen.ConnectionConfig
}

func packetgenJob(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobConfig, err := parsePacketgenArgs(ctx, logger, globalConfig, args)
	if err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	backoffController := utils.NewBackoffController(utils.NonNilBackoffConfigOrDefault(jobConfig.Backoff, globalConfig.Backoff))

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {
		err := sendPacket(ctx, logger, jobConfig, trafficMonitor)
		if err != nil {
			logger.Debug("error sending packet", zap.Error(err), zap.Any("args", args))
			utils.Sleep(ctx, backoffController.Increment().GetTimeout())
		} else {
			backoffController.Reset()
		}
	}

	return nil, nil
}

func sendPacket(ctx context.Context, logger *zap.Logger, jobConfig *packetgenJobConfig, trafficMonitor *metrics.Writer) error {
	conn, err := packetgen.OpenConnection(jobConfig.Connection)
	if err != nil {
		return err
	}

	for jobConfig.Next(ctx) {
		packetConfigRaw := jobConfig.Packet.Execute(logger, ctx)
		logger.Debug("rendered packet config template", zap.Reflect("config", packetConfigRaw))

		var packetConfig packetgen.PacketConfig
		if err := utils.Decode(packetConfigRaw, &packetConfig); err != nil {
			return err
		}

		packet, err := packetConfig.Build()
		if err != nil {
			return err
		}

		n, err := conn.Write(packet)
		if err != nil {
			return err
		}

		trafficMonitor.Add(uint64(n))
	}

	return nil
}

func parsePacketgenArgs(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (tpl *packetgenJobConfig, err error) {
	var jobConfig struct {
		BasicJobConfig
		Packet     map[string]interface{}
		Connection packetgen.ConnectionConfig
	}

	if err = ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	packetTpl, err := templates.ParseMapStruct(jobConfig.Packet)
	if err != nil {
		return nil, fmt.Errorf("error parsing packet: %w", err)
	}

	if globalConfig.ProxyURLs != "" && jobConfig.Connection.Args["protocol"] == "tcp" {
		jobConfig.Connection.Args["proxy_urls"] = templates.ParseAndExecute(logger, globalConfig.ProxyURLs, ctx)
	}

	return &packetgenJobConfig{
		BasicJobConfig: jobConfig.BasicJobConfig,
		Packet:         packetTpl,
		Connection:     jobConfig.Connection,
	}, nil
}
