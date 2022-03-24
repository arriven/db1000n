// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

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
	"math/rand"
	"net/http"
	pprofhttp "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/runner"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/ota"
	"github.com/Arriven/db1000n/src/utils/updater"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.Printf("DB1000n [Version: %s][PID=%d]\n", ota.Version, os.Getpid())

	runnerConfigOptions := runner.NewConfigOptionsWithFlags()
	jobsGlobalConfig := jobs.NewGlobalConfigWithFlags()
	otaConfig := ota.NewConfigWithFlags()
	countryCheckerConfig := utils.NewCountryCheckerConfigWithFlags()

	// Prometheus
	prometheusOn := flag.Bool("prometheus_on", utils.GetEnvBoolDefault("PROMETHEUS_ON", true),
		"Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>)")
	prometheusPushGateways := flag.String("prometheus_gateways",
		utils.GetEnvStringDefault("PROMETHEUS_GATEWAYS", "https://178.62.78.144:9091,https://46.101.26.43:9091,https://178.62.33.149:9091"),
		"Comma separated list of prometheus push gateways")

	// Config updater
	updaterMode := flag.Bool("updater-mode", utils.GetEnvBoolDefault("UPDATER_MODE", false), "Only run config updater")
	destinationConfig := flag.String("updater-destination-config", utils.GetEnvStringDefault("UPDATER_DESTINATION_CONFIG", "config/config.json"),
		"Destination config file to write (only applies if updater-mode is enabled")

	// Misc
	pprof := flag.String("pprof", utils.GetEnvStringDefault("GO_PPROF_ENDPOINT", ""), "enable pprof")
	help := flag.Bool("h", false, "print help message and exit")

	flag.Parse()

	switch {
	case *help:
		flag.CommandLine.Usage()

		return
	case *updaterMode:
		updater.Run(*destinationConfig, strings.Split(runnerConfigOptions.PathsCSV, ","), []byte(runnerConfigOptions.BackupConfig))

		return
	}

	logger, err := newZapLogger(jobsGlobalConfig.Debug)
	if err != nil {
		log.Fatalf("failed to initialize Zap logger: %v", err)
	}

	go ota.WatchUpdates(otaConfig)
	setUpPprof(*pprof, jobsGlobalConfig.Debug)
	rand.Seed(time.Now().UnixNano())

	clientID := uuid.NewString()
	country := utils.CheckCountryOrFail(countryCheckerConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initMetricsOrFail(ctx, *prometheusOn, *prometheusPushGateways, clientID, country)

	r, err := runner.New(runnerConfigOptions, jobsGlobalConfig)
	if err != nil {
		log.Panicf("Error initializing runner: %v", err)
	}

	go cancelOnSignal(cancel)
	r.Run(ctx, logger)
}

func newZapLogger(debug bool) (*zap.Logger, error) {
	if debug {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}

func setUpPprof(pprof string, debug bool) {
	if debug && pprof == "" {
		pprof = ":8080"
	}

	if pprof == "" {
		return
	}

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

func initMetricsOrFail(ctx context.Context, prometheusOn bool, prometheusPushGateways, clientID, country string) {
	if !metrics.ValidatePrometheusPushGateways(prometheusPushGateways) {
		log.Fatal("Invalid value for --prometheus_gateways")
	}

	if prometheusOn {
		metrics.InitMetrics(clientID, country)

		go metrics.ExportPrometheusMetrics(ctx, prometheusPushGateways)
	}
}

func cancelOnSignal(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
	)
	<-sigs
	log.Println("Terminating")
	cancel()
}
