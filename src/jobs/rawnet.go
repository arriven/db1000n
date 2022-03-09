package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

// rawNetJobConfig comment for linter
type rawNetJobConfig struct {
	BasicJobConfig

	Address string
	Body    json.RawMessage
}

func tcpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) error {
	defer utils.PanicHandler()

	type tcpJobConfig struct {
		rawNetJobConfig
	}

	var jobConfig tcpJobConfig
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.TrimSpace(templates.ParseAndExecute(jobConfig.Address, nil)))
	if err != nil {
		return err
	}

	bodyTpl, err := templates.Parse(string(jobConfig.Body))
	if err != nil {
		return fmt.Errorf("error parsing body template %q: %v", jobConfig.Body, err)
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
		trafficMonitor.Add(len(body))

		if err != nil {
			metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusFail)
			if debug {
				log.Printf("%s failed at %d with err: %s", jobConfig.Address, time.Now().Unix(), err.Error())
			}

		} else {
			if debug {
				log.Printf("%s finished at %d", jobConfig.Address, time.Now().Unix())
			}
			processedTrafficMonitor.Add(len(body))
			metrics.IncRawnetTCP(tcpAddr.String(), metrics.StatusSuccess)
		}
	}

	time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)

	return nil
}

func udpJob(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) error {
	defer utils.PanicHandler()

	type udpJobConfig struct {
		rawNetJobConfig
	}

	var jobConfig udpJobConfig
	if err := mapstructure.Decode(args, &jobConfig); err != nil {
		return err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", strings.TrimSpace(templates.ParseAndExecute(jobConfig.Address, nil)))
	if err != nil {
		return err
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

		return err
	}

	bodyTpl, err := templates.Parse(string(jobConfig.Body))
	if err != nil {
		return fmt.Errorf("error parsing body template %q: %v", jobConfig.Body, err)
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
			trafficMonitor.Add(len(body))
			metrics.IncRawnetUDP(udpAddr.String(), metrics.StatusSuccess)

			if debug {
				log.Printf("%s started at %d", jobConfig.Address, time.Now().Unix())
			}
		}

		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}

	return nil
}
