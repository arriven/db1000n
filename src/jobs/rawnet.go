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

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

type rawnetConfig struct {
	BasicJobConfig
	addr    string
	bodyTpl *template.Template
}

func tcpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	defer utils.PanicHandler()

	jobConfig, err := parseRawNetJobArgs(args)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	processedTrafficMonitor := metrics.Default.NewWriter(metrics.ProcessedTraffic, uuid.NewString())
	go processedTrafficMonitor.Update(ctx, time.Second)

	if debug {
		log.Printf("%s started at %d", jobConfig.addr, time.Now().Unix())
	}

	for jobConfig.Next(ctx) {
		sendTCP(ctx, jobConfig, trafficMonitor, processedTrafficMonitor, debug)
	}

	return nil, nil
}

func sendTCP(ctx context.Context, jobConfig *rawnetConfig, trafficMonitor, processedTrafficMonitor *metrics.Writer, debug bool) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", jobConfig.addr)
	if err != nil {
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		if debug {
			log.Printf("error connecting to [%v]: %v", tcpAddr, err)
		}

		metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusFail)

		return
	}

	defer conn.Close()

	// Write to conn until error
	for jobConfig.Next(ctx) {
		n, err := conn.Write([]byte(templates.Execute(jobConfig.bodyTpl, nil)))
		trafficMonitor.Add(uint64(n))

		if err != nil {
			if debug {
				log.Printf("%s failed at %d with err: %v", tcpAddr.String(), time.Now().Unix(), err)
			}

			metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusFail)

			return
		}

		if debug {
			log.Printf("%s finished at %d", tcpAddr.String(), time.Now().Unix())
		}

		processedTrafficMonitor.Add(uint64(n))
		metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusSuccess)
	}
}

func udpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	defer utils.PanicHandler()

	jobConfig, err := parseRawNetJobArgs(args)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	udpAddr, err := net.ResolveUDPAddr("udp", jobConfig.addr)
	if err != nil {
		return nil, err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	if debug {
		log.Printf("%s started at %d", jobConfig.addr, time.Now().Unix())
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		if debug {
			log.Printf("Error connecting to [%v]: %v", udpAddr, err)
		}

		metrics.IncRawnetUDP(udpAddr.String(), metrics.StatusFail)

		return nil, err
	}

	defer conn.Close()

	for jobConfig.Next(ctx) {
		sendUDP(udpAddr, conn, jobConfig.bodyTpl, trafficMonitor, debug)
	}

	return nil, nil
}

func sendUDP(a *net.UDPAddr, conn *net.UDPConn, bodyTpl *template.Template, trafficMonitor *metrics.Writer, debug bool) {
	n, err := conn.Write([]byte(templates.Execute(bodyTpl, nil)))
	if err != nil {
		metrics.IncRawnetUDP(a.String(), metrics.StatusFail)

		if debug {
			log.Printf("%s failed at %d with err: %v", a.String(), time.Now().Unix(), err)
		}

		return
	}

	trafficMonitor.Add(uint64(n))
	metrics.IncRawnetUDP(a.String(), metrics.StatusSuccess)

	if debug {
		log.Printf("%s finished at %d", a.String(), time.Now().Unix())
	}
}

func parseRawNetJobArgs(args Args) (tpl *rawnetConfig, err error) {
	var jobConfig struct {
		BasicJobConfig

		Address string
		Body    string
	}

	if err := utils.Decode(args, &jobConfig); err != nil {
		return nil, fmt.Errorf("error decoding rawnet job config: %w", err)
	}

	bodyTpl, err := templates.Parse(jobConfig.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing body template %q: %w", jobConfig.Body, err)
	}

	targetAddress := strings.TrimSpace(templates.ParseAndExecute(jobConfig.Address, nil))

	return &rawnetConfig{BasicJobConfig: jobConfig.BasicJobConfig, addr: targetAddress, bodyTpl: bodyTpl}, nil
}
