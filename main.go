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
	"log"
	"net/http"
	pprofhttp "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/runner"
	"github.com/Arriven/db1000n/src/runner/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC)

	var configPaths string
	var proxiesURL string
	var systemProxy string
	var backupConfig string
	var refreshTimeout time.Duration
	var debug, help bool
	var pprof string
	var metricsPath string
	var configFormat string
	var prometheusPushGateways string
	var prometheusOn bool
	flag.StringVar(&configPaths, "c", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json", "path to config files, separated by a comma, each path can be a web endpoint")
	flag.StringVar(&backupConfig, "b", config.DefaultConfig, "raw backup config in case the primary one is unavailable")
	flag.DurationVar(&refreshTimeout, "refresh-interval", time.Minute, "refresh timeout for updating the config")
	flag.BoolVar(&debug, "debug", false, "enable debug level logging")
	flag.StringVar(&pprof, "pprof", "", "enable pprof")
	flag.BoolVar(&help, "h", false, "print help message and exit")
	flag.StringVar(&metricsPath, "metrics-url", "", "path where to dump usage metrics, can be URL or file, empty to disable")
	flag.StringVar(&proxiesURL, "proxylist-url", "", "url to fetch proxylist")
	flag.StringVar(&systemProxy, "proxy", "", "system proxy to set by default")
	flag.StringVar(&configFormat, "format", "json", "config format")
	flag.BoolVar(&prometheusOn, "prometheus_on", false, "Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>)")
	flag.StringVar(&prometheusPushGateways, "prometheus_gateways", "", "Comma separated list of prometheus push gateways")
	flag.Parse()

	if help {
		flag.CommandLine.Usage()
		return
	}

	if debug && pprof == "" {
		pprof = ":8080"
	}

	if pprof != "" {
		mux := http.NewServeMux()
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprofhttp.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprofhttp.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprofhttp.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprofhttp.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprofhttp.Trace))
		go func() {
			log.Println(http.ListenAndServe(pprof, mux))
		}()
	}

	if !metrics.ValidatePrometheusPushGateways(prometheusPushGateways) {
		log.Fatal("Invalid value for --prometheus_gateways")
	}

	if proxiesURL != "" {
		templates.SetProxiesURL(proxiesURL)
	}

	go utils.CheckCountry([]string{"Ukraine"})

	r, err := runner.New(&runner.Config{
		ConfigPaths:        configPaths,
		BackupConfig:       []byte(backupConfig),
		RefreshTimeout:     refreshTimeout,
		MetricsPath:        metricsPath,
		Format:             configFormat,
		Global:             jobs.GlobalConfig{ProxyURL: systemProxy},
		PrometheusOn:       prometheusOn,
		PrometheusGateways: prometheusPushGateways,
	}, debug)
	if err != nil {
		log.Panicf("Error initializing runner: %v", err)
	}

	go func() {
		// Wait for sigterm
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		<-sigs
		log.Println("Terminating")
		r.Stop()
	}()

	r.Run()
}
