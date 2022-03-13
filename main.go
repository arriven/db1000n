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
	"math/rand"
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

const (
	DefaultUpdateCheckFrequency = 24 * time.Hour
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.Printf("DB1000n [Version: %s][PID=%d]\n", ota.Version, os.Getpid())

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
	doAutoUpdate := flag.Bool("enable-self-update", utils.GetEnvBoolDefault("ENABLE_SELF_UPDATE", false), "Enable the application automatic updates on the startup")
	doRestartOnUpdate := flag.Bool("restart-on-update", utils.GetEnvBoolDefault("RESTART_ON_UPDATE", true), "Allows application to restart upon successful update (ignored if auto-update is disabled)")
	skipUpdateCheckOnStart := flag.Bool("skip-update-check-on-start", utils.GetEnvBoolDefault("SKIP_UPDATE_CHECK_ON_START", false), "Allows to skip the update check at the startup (usually set automatically by the previous version)")
	autoUpdateCheckFrequency := flag.Duration("self-update-check-frequency", utils.GetEnvDurationDefault("SELF_UPDATE_CHECK_FREQUENCY", DefaultUpdateCheckFrequency), "How often to run auto-update checks")

	flag.Parse()

	if *help {
		flag.CommandLine.Usage()

		return
	}

	if *doAutoUpdate {
		go watchUpdates(*doRestartOnUpdate, *skipUpdateCheckOnStart, *autoUpdateCheckFrequency)
	}

	setUpPprof(*pprof, *debug)
	rand.Seed(time.Now().UnixNano())

	if !metrics.ValidatePrometheusPushGateways(*prometheusPushGateways) {
		log.Fatal("Invalid value for --prometheus_gateways")
	}

	if *proxiesURL != "" {
		templates.SetProxiesURL(*proxiesURL)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go utils.CheckCountry(ctx, []string{"Ukraine"})

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
	}()

	r.Run(ctx)
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

func watchUpdates(doRestartOnUpdate, skipUpdateCheckOnStart bool, autoUpdateCheckFrequency time.Duration) {
	if !skipUpdateCheckOnStart {
		runUpdate(doRestartOnUpdate)
	} else {
		log.Printf("Version update on startup is skipped, next update check is scheduled in %s",
			autoUpdateCheckFrequency)
	}

	periodicalUpdateChecker := time.NewTicker(autoUpdateCheckFrequency)
	defer periodicalUpdateChecker.Stop()

	for range periodicalUpdateChecker.C {
		runUpdate(doRestartOnUpdate)
	}
}

//nolint:nestif // The nested if linter is disabled as it would add unnecessary function splitting. This function is quite obvious
func runUpdate(doRestartOnUpdate bool) {
	log.Println("Running a check for a newer version...")

	isUpdateFound, newVersion, changeLog, err := ota.DoAutoUpdate()
	if err != nil {
		log.Printf("Auto-Update is failed: %s", err)

		return
	}

	if isUpdateFound {
		log.Printf("Newer version of the application is found [version=%s]\n", newVersion)
		log.Printf("What's new:\n%s", changeLog)

		if doRestartOnUpdate {
			log.Println("Auto restart is enabled, restarting the application to run a new version")

			additionalArgs := []string{
				"-skip-update-check-on-start",
			}

			if err = ota.Restart(additionalArgs...); err != nil {
				log.Printf("Failed to restart the application after the update to the new version: %s", err)
				log.Printf("Restart the application manually to apply changes!\n")
			}
		} else {
			log.Println("Auto restart is disabled, restart the application manually to apply changes!")
		}
	} else {
		log.Println("We are running the latest version, OK!")
	}
}
