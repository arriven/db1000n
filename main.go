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

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/job"
	"github.com/Arriven/db1000n/src/job/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/ota"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.Printf("DB1000n [Version: %s][PID=%d]\n", ota.Version, os.Getpid())

	runnerConfigOptions := job.NewConfigOptionsWithFlags()
	jobsGlobalConfig := job.NewGlobalConfigWithFlags()
	otaConfig := ota.NewConfigWithFlags()
	countryCheckerConfig := utils.NewCountryCheckerConfigWithFlags()
	updaterMode, destinationPath := config.NewUpdaterOptionsWithFlags()
	prometheusOn, prometheusPushGateways := metrics.NewOptionsWithFlags()
	pprof := flag.String("pprof", utils.GetEnvStringDefault("GO_PPROF_ENDPOINT", ""), "enable pprof")
	help := flag.Bool("h", false, "print help message and exit")
	debug := flag.Bool("debug", utils.GetEnvBoolDefault("DEBUG", false), "enable debug level logging")

	flag.Parse()

	switch {
	case *help:
		flag.CommandLine.Usage()

		return
	case *updaterMode:
		config.UpdateLocal(*destinationPath, strings.Split(runnerConfigOptions.PathsCSV, ","), []byte(runnerConfigOptions.BackupConfig))

		return
	}

	logger, err := newZapLogger(*debug)
	if err != nil {
		log.Fatalf("failed to initialize Zap logger: %v", err)
	}

	go ota.WatchUpdates(otaConfig)
	setUpPprof(*pprof, *debug)
	rand.Seed(time.Now().UnixNano())

	country := utils.CheckCountryOrFail(countryCheckerConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics.InitOrFail(ctx, *prometheusOn, *prometheusPushGateways, jobsGlobalConfig.ClientID, country)

	r, err := job.NewRunner(runnerConfigOptions, jobsGlobalConfig)
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
	switch {
	case debug && pprof == "":
		pprof = ":8080"
	case pprof == "":
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprofhttp.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprofhttp.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprofhttp.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprofhttp.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprofhttp.Trace))

	// this has to be wrapped into a lambda bc otherwise it blocks when evaluating argument for log.Println
	go func() { log.Println(http.ListenAndServe(pprof, mux)) }()
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
