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

func tcpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	defer utils.PanicHandler()

	jobConfig, addr, bodyTpl, err := parseRawNetJobArgs(args)
	if err != nil {
		return nil, err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
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
		log.Printf("%s started at %d", addr, time.Now().Unix())
	}

	for jobConfig.Next(ctx) {
		sendTCP(tcpAddr, bodyTpl, trafficMonitor, processedTrafficMonitor, debug)
	}

	return nil, nil
}

func sendTCP(a *net.TCPAddr, bodyTpl *template.Template, trafficMonitor, processedTrafficMonitor *metrics.Writer, debug bool) {
	conn, err := net.DialTCP("tcp", nil, a)
	if err != nil {
		if debug {
			log.Printf("error connecting to [%v]: %v", a, err)
		}

		metrics.IncRawnetTCP(a.String(), metrics.StatusFail)

		return
	}

	defer conn.Close()

	n, err := conn.Write([]byte(templates.Execute(bodyTpl, nil)))
	trafficMonitor.Add(uint64(n))

	if err != nil {
		if debug {
			log.Printf("%s failed at %d with err: %v", a.String(), time.Now().Unix(), err)
		}

		metrics.IncRawnetTCP(a.String(), metrics.StatusFail)

		return
	}

	if debug {
		log.Printf("%s finished at %d", a.String(), time.Now().Unix())
	}

	processedTrafficMonitor.Add(uint64(n))
	metrics.IncRawnetTCP(a.String(), metrics.StatusSuccess)
}

func udpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	defer utils.PanicHandler()

	jobConfig, addr, bodyTpl, err := parseRawNetJobArgs(args)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	if debug {
		log.Printf("%s started at %d", addr, time.Now().Unix())
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
		sendUDP(udpAddr, conn, bodyTpl, trafficMonitor, debug)
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

func parseRawNetJobArgs(args Args) (cfg *BasicJobConfig, targetAddress string, bodyTpl *template.Template, err error) {
	var jobConfig struct {
		BasicJobConfig

		Address string
		Body    string
	}

	if err := utils.Decode(args, &jobConfig); err != nil {
		return nil, "", nil, fmt.Errorf("error decoding rawnet job config: %w", err)
	}

	bodyTpl, err = templates.Parse(jobConfig.Body)
	if err != nil {
		return nil, "", nil, fmt.Errorf("error parsing body template %q: %w", jobConfig.Body, err)
	}

	targetAddress = strings.TrimSpace(templates.ParseAndExecute(jobConfig.Address, nil))

	return &jobConfig.BasicJobConfig, targetAddress, bodyTpl, nil
}
