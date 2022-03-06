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
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"

	"github.com/Arriven/db1000n/src/config"
	"github.com/Arriven/db1000n/src/logs"
	"github.com/Arriven/db1000n/src/runner"
)

func main() {
	var configPaths string
	var proxiesURL string
	var backupConfig string
	var refreshTimeout time.Duration
	var logLevel logs.Level
	var help bool
	var metricsPath string
	flag.StringVar(&configPaths, "c", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json", "path to config files, separated by a comma, each path can be a web endpoint")
	flag.StringVar(&backupConfig, "b", config.DefaultConfig, "raw backup config in case the primary one is unavailable")
	flag.DurationVar(&refreshTimeout, "r", time.Minute, "refresh timeout for updating the config")
	flag.IntVar(&logLevel, "l", logs.Info, "logging level. 0 - Debug, 1 - Info, 2 - Warning, 3 - Error")
	flag.BoolVar(&help, "h", false, "print help message and exit")
	flag.StringVar(&metricsPath, "m", "", "path where to dump usage metrics, can be URL or file, empty to disable")
	flag.StringVar(&proxiesURL, "p", "", "url to fetch proxies list")
	flag.Parse()

	if help {
		flag.CommandLine.Usage()
		return
	}

	logs.Default = logs.New(logLevel)

	if proxiesURL != "" {
		templates.SetProxiesUrl(proxiesURL)
	}

	go utils.CheckCountry(logs.Default, []string{"Ukraine"})

	r, err := runner.New(&runner.Config{
		ConfigPaths:    configPaths,
		BackupConfig:   []byte(backupConfig),
		RefreshTimeout: refreshTimeout,
		MetricsPath:    metricsPath,
	}, logs.Default)
	if err != nil {
		logs.Default.Error("Error initializing runner: %v", err)
	}

	go func() {
		// Wait for sigterm
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		<-sigs
		logs.Default.Info("Terminating")
		r.Stop()
	}()

	r.Run()
}
