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
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/packetgen"
	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func packetgenJob(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var jobConfig struct {
		BasicJobConfig
		Packet     map[string]interface{}
		Connection packetgen.ConnectionConfig
	}

	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	rawConn, err := packetgen.OpenRawConnection(jobConfig.Connection)
	if err != nil {
		return nil, fmt.Errorf("error building raw connection: %w", err)
	}

	packetTpl, err := templates.ParseMapStruct(jobConfig.Packet)
	if err != nil {
		return nil, fmt.Errorf("error parsing packet: %w", err)
	}

	payloadBuf := gopacket.NewSerializeBuffer()

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {
		if err := payloadBuf.Clear(); err != nil {
			return nil, fmt.Errorf("error clearing payload buffer: %w", err)
		}

		packetConfigRaw := packetTpl.Execute(logger, ctx)
		logger.Debug("rendered packet config template", zap.Reflect("config", packetConfigRaw))

		var packetConfig packetgen.PacketConfig
		if err := utils.Decode(packetConfigRaw, &packetConfig); err != nil {
			return nil, fmt.Errorf("error parsing packet config: %w", err)
		}

		packet, err := packetConfig.Build()
		if err != nil {
			return nil, fmt.Errorf("error building packet: %w", err)
		}

		if err = packet.Serialize(payloadBuf); err != nil {
			return nil, fmt.Errorf("error serializing packet: %w", err)
		}

		n, err := rawConn.WriteTo(payloadBuf.Bytes(), nil, &net.IPAddr{IP: packet.IP()})
		if err != nil {
			return nil, fmt.Errorf("error sending packet: %w", err)
		}

		trafficMonitor.Add(uint64(n))
	}

	return nil, nil
}
