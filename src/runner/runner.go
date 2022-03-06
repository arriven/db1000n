package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/logs"
	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/utils"
)

// Config for the job runner
type Config struct {
	ConfigPaths    string        // Comma-separated config location URLs
	BackupConfig   []byte        // Raw backup config
	RefreshTimeout time.Duration // How often to refresh config
	MetricsPath    string        // Where to dump metrics to
}

// Runner executes jobs according to the (fetched from remote) configuration
type Runner struct {
	configPaths    []string
	backupConfig   []byte
	refreshTimeout time.Duration
	metricsPath    string

	currentRawConfig []byte // currently applied config

	log *logs.Logger

	stop chan interface{}
}

// New runner according to the config
func New(cfg *Config, l *logs.Logger) (*Runner, error) {
	return &Runner{
		configPaths:    strings.Split(cfg.ConfigPaths, ","),
		backupConfig:   cfg.BackupConfig,
		refreshTimeout: cfg.RefreshTimeout,
		metricsPath:    cfg.MetricsPath,

		log:  l,
		stop: make(chan interface{}),
	}, nil
}

// Run the runner and block until Stop() is called
func (r *Runner) Run() {
	clientID := uuid.New().String()
	refreshTimer := time.NewTicker(r.refreshTimeout)

	var (
		stop   bool
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup
	)

	for !stop {
		newRawConfig, err := fetchConfig(r.configPaths, r.log)
		if err != nil {
			if r.currentRawConfig != nil {
				r.log.Warning("Could not load new config, proceeding with last known good config")
				newRawConfig = r.currentRawConfig
			} else {
				r.log.Warning("Could not load new config, proceeding with backupConfig")
				newRawConfig = r.backupConfig
			}
		}

		if !bytes.Equal(r.currentRawConfig, newRawConfig) { // Only restart jobs if the new config differs from the current one
			r.log.Info("New config received, applying")

			var config struct {
				Jobs []jobs.Config
			}

			if err := json.Unmarshal(newRawConfig, &config); err != nil {
				r.log.Warning("Failed to unmarshal job configs: %v", err)
			} else {
				if cancel != nil {
					cancel()
				}

				ctx, cancel = context.WithCancel(context.Background())

				for i := range config.Jobs {
					job, ok := jobs.Get(config.Jobs[i].Type)
					if !ok {
						r.log.Warning("No such job %q", config.Jobs[i].Type)

						continue
					}

					if config.Jobs[i].Count < 1 {
						config.Jobs[i].Count = 1
					}

					for j := 0; j < config.Jobs[i].Count; j++ {
						wg.Add(1)

						go func(i int) {
							job(ctx, r.log, config.Jobs[i].Args)
							wg.Done()
						}(i)
					}
				}

				r.currentRawConfig = newRawConfig

				r.log.Info("Jobs (re)started")
			}
		}

		// Wait for refresh timer or stop signal
		select {
		case <-refreshTimer.C:
		case <-r.stop:
			refreshTimer.Stop()

			stop = true
		}

		dumpMetrics(r.log, r.metricsPath, "traffic", clientID)
	}

	if cancel != nil {
		cancel()
	}

	wg.Wait()
}

// Stop runner asynchronously
func (r *Runner) Stop() { close(r.stop) }

func fetchConfig(paths []string, l *logs.Logger) ([]byte, error) {
	for i := range paths {
		res, err := fetchSingleConfig(paths[i])
		if err != nil {
			l.Warning("Failed to fetch config from %q: %v", paths[i], err)
			continue
		}

		l.Info("Loading config from %q", paths[i])

		return res, nil
	}

	return nil, errors.New("config fetch failed")
}

func fetchSingleConfig(path string) ([]byte, error) {
	configURL, err := url.ParseRequestURI(path)
	if err != nil {
		res, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	resp, err := http.Get(configURL.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error fetching config, code %d", resp.StatusCode)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func dumpMetrics(l *logs.Logger, path, name, clientID string) {
	defer func() {
		if err := recover(); err != nil {
			logs.Default.Warning("caught panic: %v", err)
		}
	}()

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
