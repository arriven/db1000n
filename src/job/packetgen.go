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
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/packetgen"
	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

type packetgenJobConfig struct {
	BasicJobConfig
	Dynamic    bool
	Packets    []*templates.MapStruct
	Connection packetgen.ConnectionConfig
}

// "packetgen" in config
func packetgenJob(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (data any, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobConfig, err := parsePacketgenArgs(ctx, args, globalConfig, a, logger)
	if err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	backoffController := utils.BackoffController{BackoffConfig: utils.NonNilOrDefault(jobConfig.Backoff, globalConfig.Backoff)}

	for jobConfig.Next(ctx) {
		if err := sendPacket(ctx, logger, jobConfig, a); err != nil {
			logger.Debug("error sending packet", zap.Error(err), zap.Any("args", args))
			utils.Sleep(ctx, backoffController.Increment().GetTimeout())
		} else {
			backoffController.Reset()
		}
	}

	return nil, nil
}

func sendPacket(ctx context.Context, logger *zap.Logger, jobConfig *packetgenJobConfig, a *metrics.Accumulator) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := packetgen.OpenConnection(ctx, jobConfig.Connection)
	if err != nil {
		return err
	}
	defer conn.Close()

	packetSrc, err := makePacketSource(ctx, logger, jobConfig.Packets, jobConfig.Dynamic)
	if err != nil {
		return err
	}

	for jobConfig.Next(ctx) {
		packet, err := packetSrc(ctx, logger)
		if err != nil {
			return err
		}

		n, err := conn.Write(packet)
		if err != nil {
			if a != nil {
				a.Inc(conn.Target(), metrics.RequestsAttemptedStat).Flush()
			}

			return err
		}

		if a != nil {
			tgt := conn.Target()

			a.Inc(tgt, metrics.RequestsAttemptedStat).
				Inc(tgt, metrics.RequestsSentStat).
				Add(tgt, metrics.BytesSentStat, uint64(n)).
				Flush()
		}
	}

	return nil
}

type packetSource func(ctx context.Context, logger *zap.Logger) (packetgen.Packet, error)

func makePacketSource(ctx context.Context, logger *zap.Logger, packetTpls []*templates.MapStruct, dynamic bool) (packetSource, error) {
	if dynamic {
		packetsChan := utils.InfiniteRange(ctx, packetTpls)

		return func(ctx context.Context, logger *zap.Logger) (packetgen.Packet, error) {
			return getNextDynamicPacket(ctx, logger, packetsChan)
		}, nil
	}

	packets, err := staticPackets(ctx, logger, packetTpls)
	if err != nil {
		return nil, err
	}

	packetsChan := utils.InfiniteRange(ctx, packets)

	return func(ctx context.Context, logger *zap.Logger) (packetgen.Packet, error) {
		return getNextStaticPacket(ctx, logger, packetsChan)
	}, nil
}

func staticPackets(ctx context.Context, logger *zap.Logger, packetTpls []*templates.MapStruct) ([]packetgen.Packet, error) {
	packets := make([]packetgen.Packet, 0, len(packetTpls))

	for _, packetTpl := range packetTpls {
		packetConfigRaw := packetTpl.Execute(logger, ctx)

		var packetConfig packetgen.PacketConfig
		if err := utils.Decode(packetConfigRaw, &packetConfig); err != nil {
			return nil, err
		}

		packet, err := packetConfig.Build()
		if err != nil {
			return nil, err
		}

		packets = append(packets, packet)
	}

	return packets, nil
}

func getNextStaticPacket(ctx context.Context, logger *zap.Logger, packetsChan chan packetgen.Packet) (packetgen.Packet, error) {
	select {
	case <-ctx.Done():
		return packetgen.Packet{}, ctx.Err()
	case packet, more := <-packetsChan:
		if !more {
			return packetgen.Packet{}, errors.New("packetsChan closed")
		}

		return packet, nil
	}
}

func getNextDynamicPacket(ctx context.Context, logger *zap.Logger, packetsChan chan *templates.MapStruct) (packetgen.Packet, error) {
	select {
	case <-ctx.Done():
		return packetgen.Packet{}, ctx.Err()
	case packetTpl, more := <-packetsChan:
		if !more {
			return packetgen.Packet{}, errors.New("packetsChan closed")
		}

		packetConfigRaw := packetTpl.Execute(logger, ctx)

		var packetConfig packetgen.PacketConfig
		if err := utils.Decode(packetConfigRaw, &packetConfig); err != nil {
			return packetgen.Packet{}, err
		}

		return packetConfig.Build()
	}
}

type packetDescriptor struct {
	Packet map[string]any
	Count  int
}

func parsePackets(packets []packetDescriptor, dflt map[string]any) ([]*templates.MapStruct, error) {
	packetTpls := make([]*templates.MapStruct, 0, len(packets)+1)

	if len(packets) > 0 {
		for _, descriptor := range packets {
			if descriptor.Count < 1 {
				descriptor.Count = 1
			}

			packetTpl, err := templates.ParseMapStruct(descriptor.Packet)
			if err != nil {
				return nil, fmt.Errorf("error parsing packet: %w", err)
			}

			for i := 0; i < descriptor.Count; i++ {
				packetTpls = append(packetTpls, packetTpl)
			}
		}
	} else {
		packetTpl, err := templates.ParseMapStruct(dflt)
		if err != nil {
			return nil, fmt.Errorf("error parsing packet: %w", err)
		}

		packetTpls = append(packetTpls, packetTpl)
	}

	return packetTpls, nil
}

func parsePacketgenArgs(ctx context.Context, args config.Args, globalConfig *GlobalConfig, a *metrics.Accumulator, logger *zap.Logger) (
	tpl *packetgenJobConfig, err error,
) {
	var jobConfig struct {
		BasicJobConfig
		Dynamic    bool
		Packet     map[string]any
		Packets    []packetDescriptor
		Connection packetgen.ConnectionConfig
	}

	if err = ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing job config: %w", err)
	}

	packetTpls, err := parsePackets(jobConfig.Packets, jobConfig.Packet)
	if err != nil {
		return nil, err
	}

	proxyCfg := utils.NonNilOrDefault(jobConfig.Connection.Proxy, globalConfig.GetProxyParams(logger, ctx))
	jobConfig.Connection.Proxy = &proxyCfg

	return &packetgenJobConfig{
		BasicJobConfig: jobConfig.BasicJobConfig,
		Dynamic:        jobConfig.Dynamic,
		Packets:        packetTpls,
		Connection:     jobConfig.Connection,
	}, nil
}
