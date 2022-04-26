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
	"github.com/Arriven/db1000n/src/utils/templates"
)

const simpleLogFormat = "simple"

func main() {
	runnerConfigOptions := job.NewConfigOptionsWithFlags()
	jobsGlobalConfig := job.NewGlobalConfigWithFlags()
	otaConfig := ota.NewConfigWithFlags()
	countryCheckerConfig := utils.NewCountryCheckerConfigWithFlags()
	updaterMode, destinationPath := config.NewUpdaterOptionsWithFlags()
	prometheusOn, prometheusListenAddress, prometheusPushGateways := metrics.NewOptionsWithFlags()
	pprof := flag.String("pprof", utils.GetEnvStringDefault("GO_PPROF_ENDPOINT", ""), "enable pprof")
	help := flag.Bool("h", false, "print help message and exit")
	version := flag.Bool("version", false, "print version and exit")
	debug := flag.Bool("debug", utils.GetEnvBoolDefault("DEBUG", false), "enable debug level logging and features")
	logLevel := flag.String("log-level", utils.GetEnvStringDefault("LOG_LEVEL", "none"), "log level override for zap, leave empty to use default")
	logFormat := flag.String("log-format", utils.GetEnvStringDefault("LOG_FORMAT", simpleLogFormat), "overrides the default (simple) log output format,\n"+
		"possible values are: json, console, simple\n"+
		"simple is the most human readable format if you only look at the output in your terminal")

	flag.Parse()

	logger, err := newZapLogger(*debug, *logLevel, *logFormat)
	if err != nil {
		panic(err)
	}

	logger.Info("running db1000n", zap.String("version", ota.Version), zap.Int("pid", os.Getpid()))

	switch {
	case *help:
		flag.CommandLine.Usage()

		return
	case *version:
		return
	case *updaterMode:
		config.UpdateLocal(logger, *destinationPath, strings.Split(runnerConfigOptions.PathsCSV, ","), []byte(runnerConfigOptions.BackupConfig))

		return
	}

	err = utils.UpdateRLimit(logger)
	if err != nil {
		logger.Warn("failed to increase rlimit", zap.Error(err))
	}

	go ota.WatchUpdates(logger, otaConfig)
	setUpPprof(logger, *pprof, *debug)
	rand.Seed(time.Now().UnixNano())

	country := utils.CheckCountryOrFail(logger, countryCheckerConfig, templates.ParseAndExecute(logger, jobsGlobalConfig.ProxyURLs, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics.InitOrFail(ctx, logger, *prometheusOn, *prometheusListenAddress, *prometheusPushGateways, jobsGlobalConfig.ClientID, country)

	go cancelOnSignal(logger, cancel)

	reporter := newReporter(*logFormat, logger)
	job.NewRunner(runnerConfigOptions, jobsGlobalConfig, reporter).Run(ctx, logger)
}

func newZapLogger(debug bool, logLevel string, logFormat string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if debug {
		cfg = zap.NewDevelopmentConfig()
	}

	if logFormat == simpleLogFormat {
		// turn off all output except the message itself
		cfg.Encoding = "console"
		cfg.EncoderConfig.LevelKey = ""
		cfg.EncoderConfig.TimeKey = ""
		cfg.EncoderConfig.NameKey = ""

		// turning these off for debug output would be undesirable
		if !debug {
			cfg.EncoderConfig.CallerKey = ""
			cfg.EncoderConfig.StacktraceKey = ""
		}
	} else if logFormat != "" {
		cfg.Encoding = logFormat
	}

	level, err := zap.ParseAtomicLevel(logLevel)
	if err == nil {
		cfg.Level = level
	}

	return cfg.Build()
}

func setUpPprof(logger *zap.Logger, pprof string, debug bool) {
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

	// this has to be wrapped into a lambda bc otherwise it blocks when evaluating argument for zap.Error
	go func() { logger.Warn("pprof server", zap.Error(http.ListenAndServe(pprof, mux))) }()
}

func cancelOnSignal(logger *zap.Logger, cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGHUP,
		syscall.SIGINT,
	)
	<-sigs
	logger.Info("terminating")
	cancel()
}

func newReporter(logFormat string, logger *zap.Logger) metrics.Reporter {
	if logFormat == simpleLogFormat {
		return metrics.NewConsoleReporter(os.Stdout)
	}

	return metrics.NewZapReporter(logger)
}
