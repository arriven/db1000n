package jobs

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

// RawNetJobConfig comment for linter
type RawNetJobConfig struct {
	BasicJobConfig
	Address string
	Body    json.RawMessage
}

func tcpJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	type tcpJobConfig struct {
		RawNetJobConfig
	}

	var jobConfig tcpJobConfig
	if err := json.Unmarshal(args, &jobConfig); err != nil {
		return err
	}

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.TrimSpace(templates.Execute(jobConfig.Address)))
	if err != nil {
		return err
	}

	for jobConfig.Next(ctx) {
		if debug {
			log.Printf("%s started at %d", jobConfig.Address, time.Now().Unix())
		}

		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			if debug {
				log.Printf("error connecting to [%v]: %v", tcpAddr, err)
			}

			continue
		}

		_, err = conn.Write([]byte(templates.Execute(string(jobConfig.Body))))
		trafficMonitor.Add(len(jobConfig.Body))

		if debug {
			if err != nil {
				log.Printf("%s failed at %d with err: %s", jobConfig.Address, time.Now().Unix(), err.Error())
			} else {
				log.Printf("%s finished at %d", jobConfig.Address, time.Now().Unix())
			}
		}

		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}
	return nil
}

func udpJob(ctx context.Context, args Args, debug bool) error {
	defer utils.PanicHandler()

	type udpJobConfig struct {
		RawNetJobConfig
	}

	var jobConfig udpJobConfig
	if err := json.Unmarshal(args, &jobConfig); err != nil {
		return err
	}

	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	udpAddr, err := net.ResolveUDPAddr("udp", strings.TrimSpace(templates.Execute(jobConfig.Address)))
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

		return err
	}

	for jobConfig.Next(ctx) {
		_, err = conn.Write([]byte(templates.Execute(string(jobConfig.Body))))
		trafficMonitor.Add(len(jobConfig.Body))

		if debug {
			if err != nil {
				log.Printf("%s failed at %d with err: %s", jobConfig.Address, time.Now().Unix(), err.Error())
			} else {
				log.Printf("%s started at %d", jobConfig.Address, time.Now().Unix())
			}
		}

		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}

	return nil
}
