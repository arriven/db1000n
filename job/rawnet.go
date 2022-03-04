package job

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/Arriven/db1000n/logs"
	"github.com/Arriven/db1000n/metrics"
	"github.com/Arriven/db1000n/template"
	"github.com/google/uuid"
)

// RawNetJobConfig comment for linter
type RawNetJobConfig struct {
	BasicJobConfig
	Address string
	Body    json.RawMessage
}

func tcpJob(ctx context.Context, l *logs.Logger, args Args) error {
	defer panicHandler()
	type tcpJobConfig struct {
		RawNetJobConfig
	}
	var jobConfig tcpJobConfig
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}
	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	tcpAddr, err := net.ResolveTCPAddr("tcp", strings.TrimSpace(template.Execute(jobConfig.Address)))
	if err != nil {
		return err
	}
	for jobConfig.Next(ctx) {
		startedAt := time.Now().Unix()
		l.Debug("%s started at %d", jobConfig.Address, startedAt)

		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			l.Debug("error connecting to [%v]: %v", tcpAddr, err)
			continue
		}

		_, err = conn.Write([]byte(template.Execute(string(jobConfig.Body))))
		trafficMonitor.Add(len(jobConfig.Body))

		finishedAt := time.Now().Unix()
		if err != nil {
			l.Debug("%s failed at %d with err: %s", jobConfig.Address, finishedAt, err.Error())
		} else {
			l.Debug("%s started at %d", jobConfig.Address, finishedAt)
		}
		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}
	return nil
}

func udpJob(ctx context.Context, l *logs.Logger, args Args) error {
	defer panicHandler()
	type udpJobConfig struct {
		RawNetJobConfig
	}
	var jobConfig udpJobConfig
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}
	trafficMonitor := metrics.Default.NewWriter(ctx, "traffic", uuid.New().String())
	udpAddr, err := net.ResolveUDPAddr("udp", strings.TrimSpace(template.Execute(jobConfig.Address)))
	if err != nil {
		return err
	}
	startedAt := time.Now().Unix()
	l.Debug("%s started at %d", jobConfig.Address, startedAt)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		l.Debug("error connecting to [%v]: %v", udpAddr, err)
		return err
	}

	for jobConfig.Next(ctx) {
		_, err = conn.Write([]byte(template.Execute(string(jobConfig.Body))))
		trafficMonitor.Add(len(jobConfig.Body))

		finishedAt := time.Now().Unix()
		if err != nil {
			l.Debug("%s failed at %d with err: %s", jobConfig.Address, finishedAt, err.Error())
		} else {
			l.Debug("%s started at %d", jobConfig.Address, finishedAt)
		}
		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}
	return nil
}
