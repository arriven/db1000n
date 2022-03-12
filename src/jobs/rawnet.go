package jobs

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

// rawNetJobConfig comment for linter
type rawNetJobConfig struct {
	BasicJobConfig

	Address string
	Body    string
}

func tcpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	type tcpJobConfig struct {
		rawNetJobConfig
	}

	var jobConfig tcpJobConfig
	if err := utils.Decode(args, &jobConfig); err != nil {
		return nil, err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.TrimSpace(templates.ParseAndExecute(jobConfig.Address, nil)))
	if err != nil {
		return nil, err
	}

	bodyTpl, err := templates.Parse(jobConfig.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing body template %q: %w", jobConfig.Body, err)
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)
	processedTrafficMonitor := metrics.Default.NewWriter(metrics.ProcessedTraffic, uuid.NewString())
	go processedTrafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {
		if debug {
			log.Printf("%s started at %d", jobConfig.Address, time.Now().Unix())
		}

		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			if debug {
				log.Printf("error connecting to [%v]: %v", tcpAddr, err)
			}
			metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusFail)

			continue
		}

		body := []byte(templates.Execute(bodyTpl, nil))
		_, err = conn.Write(body)
		trafficMonitor.Add(uint64(len(body)))

		if err != nil {
			metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusFail)
			if debug {
				log.Printf("%s failed at %d with err: %s", jobConfig.Address, time.Now().Unix(), err.Error())
			}
		} else {
			if debug {
				log.Printf("%s finished at %d", jobConfig.Address, time.Now().Unix())
			}
			processedTrafficMonitor.Add(uint64(len(body)))
			metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusSuccess)
		}
	}

	return nil, nil
}

func udpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer utils.PanicHandler()

	type udpJobConfig struct {
		rawNetJobConfig
	}

	var jobConfig udpJobConfig
	if err := utils.Decode(args, &jobConfig); err != nil {
		return nil, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", strings.TrimSpace(templates.ParseAndExecute(jobConfig.Address, nil)))
	if err != nil {
		return nil, err
	}

	if debug {
		log.Printf("%s started at %d", jobConfig.Address, time.Now().Unix())
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		if debug {
			log.Printf("Error connecting to [%v]: %v", udpAddr, err)
		}
		metrics.IncRawnetUDP(udpAddr.String(), metrics.StatusFail)

		return nil, err
	}

	bodyTpl, err := templates.Parse(jobConfig.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing body template %q: %w", jobConfig.Body, err)
	}

	trafficMonitor := metrics.Default.NewWriter(metrics.Traffic, uuid.New().String())
	go trafficMonitor.Update(ctx, time.Second)

	for jobConfig.Next(ctx) {
		body := []byte(templates.Execute(bodyTpl, nil))
		if _, err = conn.Write(body); err != nil {
			metrics.IncRawnetUDP(udpAddr.String(), metrics.StatusFail)

			if debug {
				log.Printf("%s failed at %d with err: %s", jobConfig.Address, time.Now().Unix(), err.Error())
			}
		} else {
			trafficMonitor.Add(uint64(len(body)))
			metrics.IncRawnetUDP(udpAddr.String(), metrics.StatusSuccess)

			if debug {
				log.Printf("%s started at %d", jobConfig.Address, time.Now().Unix())
			}
		}
	}

	return nil, nil
}
