package runner

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/metrics"
	"github.com/Arriven/db1000n/src/runner/config"
	"github.com/Arriven/db1000n/src/utils"
	"github.com/Arriven/db1000n/src/utils/templates"
)

// Config for the job runner
type Config struct {
	ConfigPaths        string        // Comma-separated config location URLs
	BackupConfig       []byte        // Raw backup config
	RefreshTimeout     time.Duration // How often to refresh config
	MetricsPath        string        // Where to dump metrics to
	Format             string        // json or yaml
	PrometheusOn       bool
	PrometheusGateways string
}

// Runner executes jobs according to the (fetched from remote) configuration
type Runner struct {
	config         *Config
	configPaths    []string
	backupConfig   []byte
	refreshTimeout time.Duration
	metricsPath    string
	configFormat   string

	currentRawConfig []byte // currently applied config

	debug bool

	stop chan interface{}
}

// New runner according to the config
func New(cfg *Config, debug bool) (*Runner, error) {
	return &Runner{
		config:         cfg,
		configPaths:    strings.Split(cfg.ConfigPaths, ","),
		backupConfig:   cfg.BackupConfig,
		refreshTimeout: cfg.RefreshTimeout,
		metricsPath:    cfg.MetricsPath,
		configFormat:   cfg.Format,

		debug: debug,

		stop: make(chan interface{}),
	}, nil
}

// Run the runner and block until Stop() is called
func (r *Runner) Run() {
	clientID := uuid.New()
	refreshTimer := time.NewTicker(r.refreshTimeout)

	var (
		stop   bool
		ctx    context.Context
		cancel context.CancelFunc
		wg     sync.WaitGroup
	)

	for !stop {
		if cfg, raw := config.Update(r.configPaths, r.currentRawConfig, r.backupConfig, r.configFormat); cfg != nil {
			if cancel != nil {
				cancel()
			}

			ctx, cancel = context.WithCancel(context.Background())
			if r.config.PrometheusOn {
				go metrics.ExportPrometheusMetrics(ctx, r.config.PrometheusGateways)
			}

			for i := range cfg.Jobs {
				if len(cfg.Jobs[i].Filter) != 0 && strings.TrimSpace(templates.ParseAndExecute(cfg.Jobs[i].Filter, clientID.ID())) != "true" {
					log.Println("There is a filter defined for a job but this client doesn't pass it - skip the job")
					continue
				}
				job, ok := jobs.Get(cfg.Jobs[i].Type)
				if !ok {
					log.Printf("Unknown job %q", cfg.Jobs[i].Type)

					continue
				}

				if cfg.Jobs[i].Count < 1 {
					cfg.Jobs[i].Count = 1
				}

				for j := 0; j < cfg.Jobs[i].Count; j++ {
					wg.Add(1)

					go func(i int) {
						job(ctx, cfg.Jobs[i].Args, r.debug)
						wg.Done()
					}(i)
				}
			}

			r.currentRawConfig = raw

			log.Println("Jobs (re)started")
		}

		// Wait for refresh timer or stop signal
		select {
		case <-refreshTimer.C:
		case <-r.stop:
			refreshTimer.Stop()

			stop = true
		}

		dumpMetrics(r.metricsPath, "traffic", clientID.String())
	}

	if cancel != nil {
		cancel()
	}

	wg.Wait()
}

// Stop runner asynchronously
func (r *Runner) Stop() { close(r.stop) }

func dumpMetrics(path, name, clientID string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("caught panic: %v", err)
		}
	}()

	bytesGenerated := metrics.Default.Read(name)
	utils.ReportStatistics(int64(bytesGenerated), clientID)
	if bytesGenerated > 0 {
		log.Println("Атака проводиться успішно! Руський воєнний корабль іди нахуй!")
		log.Println("Attack is successful! Russian warship, go fuck yourself!")
		log.Printf("The app has generated approximately %v bytes of traffic", bytesGenerated)
	} else {
		log.Println("The app doesn't seem to generate any traffic, please contact your admin")
	}
}
