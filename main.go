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
	"github.com/Arriven/db1000n/src/runner"
	"github.com/Arriven/db1000n/src/runner/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/ota"
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
	var scaleFactor int
	var debug, help bool
	var pprof string
	var metricsPath string
	var configFormat string
	var prometheusPushGateways string
	var prometheusOn bool
	var doSelfUpdate bool

	flag.StringVar(&configPaths, "c", utils.GetEnvStringDefault("CONFIG", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.v0.7.json"), "path to config files, separated by a comma, each path can be a web endpoint")
	flag.StringVar(&backupConfig, "b", config.DefaultConfig, "raw backup config in case the primary one is unavailable")
	flag.DurationVar(&refreshTimeout, "refresh-interval", utils.GetEnvDurationDefault("REFRESH_INTERVAL", time.Minute), "refresh timeout for updating the config")
	flag.IntVar(&scaleFactor, "scale", utils.GetEnvIntDefault("SCALE_FACTOR", 1), "used to scale the amount of jobs being launched, effect is similar to launching multiple instances at once")
	flag.BoolVar(&debug, "debug", utils.GetEnvBoolDefault("DEBUG", false), "enable debug level logging")
	flag.StringVar(&pprof, "pprof", utils.GetEnvStringDefault("GO_PPROF_ENDPOINT", ""), "enable pprof")
	flag.BoolVar(&help, "h", false, "print help message and exit")
	flag.StringVar(&metricsPath, "metrics-url", utils.GetEnvStringDefault("METRICS_URL", ""), "path where to dump usage metrics, can be URL or file, empty to disable")
	flag.StringVar(&proxiesURL, "proxylist-url", utils.GetEnvStringDefault("PROXYLIST_URL", ""), "url to fetch proxylist")
	flag.StringVar(&systemProxy, "proxy", utils.GetEnvStringDefault("SYSTEM_PROXY", ""), "system proxy to set by default")
	flag.StringVar(&configFormat, "format", utils.GetEnvStringDefault("CONFIG_FORMAT", "json"), "config format")
	flag.BoolVar(&prometheusOn, "prometheus_on", utils.GetEnvBoolDefault("PROMETHEUS_ON", false), "Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>)")
	flag.StringVar(&prometheusPushGateways, "prometheus_gateways", utils.GetEnvStringDefault("PROMETHEUS_GATEWAYS", ""), "Comma separated list of prometheus push gateways")
	flag.BoolVar(&doSelfUpdate, "enable-self-update", utils.GetEnvBoolDefault("ENABLE_SELF_UPDATE", false), "Enable the application automatic updates on the startup")
	flag.Parse()

	log.Printf("DB1000n [Version: %s]\n", ota.Version)

	if help {
		flag.CommandLine.Usage()
		return
	}

	if doSelfUpdate {
		ota.DoSelfUpdate()
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
		Global:             jobs.GlobalConfig{ProxyURL: systemProxy, ScaleFactor: scaleFactor},
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
