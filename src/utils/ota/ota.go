// Package ota [allows hot update and reload of the executable]
package ota

import (
	"flag"
	"fmt"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
)

var (
	// Version is a release version embedded into the app
	Version = "v0.0.1"
	// Repository to check for updates
	Repository = "Arriven/db1000n" // Could be changed via the ldflags
)

// Config defines OTA parameters.
type Config struct {
	doAutoUpdate, doRestartOnUpdate, skipUpdateCheckOnStart bool
	autoUpdateCheckFrequency                                time.Duration
}

// NewConfigWithFlags returns a Config initialized with command line flags.
func NewConfigWithFlags() *Config {
	const defaultUpdateCheckFrequency = 24 * time.Hour

	var res Config

	flag.BoolVar(&res.doAutoUpdate, "enable-self-update", utils.GetEnvBoolDefault("ENABLE_SELF_UPDATE", false),
		"Enable the application automatic updates on the startup")
	flag.BoolVar(&res.doRestartOnUpdate, "restart-on-update", utils.GetEnvBoolDefault("RESTART_ON_UPDATE", true),
		"Allows application to restart upon successful update (ignored if auto-update is disabled)")
	flag.BoolVar(&res.skipUpdateCheckOnStart, "skip-update-check-on-start", utils.GetEnvBoolDefault("SKIP_UPDATE_CHECK_ON_START", false),
		"Allows to skip the update check at the startup (usually set automatically by the previous version)")
	flag.DurationVar(&res.autoUpdateCheckFrequency, "self-update-check-frequency",
		utils.GetEnvDurationDefault("SELF_UPDATE_CHECK_FREQUENCY", defaultUpdateCheckFrequency), "How often to run auto-update checks")

	return &res
}

// WatchUpdates performs OTA updates based on the config.
func WatchUpdates(logger *zap.Logger, cfg *Config) {
	if !cfg.doAutoUpdate {
		return
	}

	if !cfg.skipUpdateCheckOnStart {
		runUpdate(logger, cfg.doRestartOnUpdate)
	} else {
		logger.Info("version update on startup is skipped",
			zap.Duration("auto_update_check_frequency", cfg.autoUpdateCheckFrequency))
	}

	periodicalUpdateChecker := time.NewTicker(cfg.autoUpdateCheckFrequency)
	defer periodicalUpdateChecker.Stop()

	for range periodicalUpdateChecker.C {
		runUpdate(logger, cfg.doRestartOnUpdate)
	}
}

func runUpdate(logger *zap.Logger, doRestartOnUpdate bool) {
	logger.Info("running a check for a newer version")

	isUpdateFound, newVersion, changeLog, err := doAutoUpdate()

	switch {
	case err != nil:
		logger.Warn("auto-update failed", zap.Error(err))

		return
	case !isUpdateFound:
		logger.Info("running the latest version")

		return
	}

	logger.Info("newer version of the application is found", zap.String("version", newVersion))
	logger.Info("changelog", zap.String("changes", changeLog))

	if !doRestartOnUpdate {
		logger.Warn("auto restart is disabled, restart the application manually to apply changes")

		return
	}

	logger.Info("auto restart is enabled, restarting the application to run a new version")

	if err = restart(logger, "-skip-update-check-on-start"); err != nil {
		logger.Warn("Failed to restart the application after the update to the new version", zap.Error(err))
		logger.Warn("restart the application manually to apply changes")
	}
}

// doAutoUpdate updates the app to the latest version.
func doAutoUpdate() (updateFound bool, newVersion, changeLog string, err error) {
	v, err := semver.ParseTolerant(Version)
	if err != nil {
		return false, "", "", fmt.Errorf("binary version validation failed: %w", err)
	}

	latest, err := selfupdate.UpdateSelf(v, Repository)

	switch {
	case err != nil:
		return false, "", "", fmt.Errorf("binary update failed: %w", err)
	case latest.Version.Equals(v):
		return false, "", "", nil
	}

	return true, latest.Version.String(), latest.ReleaseNotes, nil
}
