// Package ota [allows hot update and reload of the executable]
package ota

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"

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
func WatchUpdates(cfg *Config) {
	if !cfg.doAutoUpdate {
		return
	}

	if !cfg.skipUpdateCheckOnStart {
		runUpdate(cfg.doRestartOnUpdate)
	} else {
		log.Printf("Version update on startup is skipped, next update check scheduled in %v",
			cfg.autoUpdateCheckFrequency)
	}

	periodicalUpdateChecker := time.NewTicker(cfg.autoUpdateCheckFrequency)
	defer periodicalUpdateChecker.Stop()

	for range periodicalUpdateChecker.C {
		runUpdate(cfg.doRestartOnUpdate)
	}
}

func runUpdate(doRestartOnUpdate bool) {
	log.Println("Running a check for a newer version...")

	isUpdateFound, newVersion, changeLog, err := doAutoUpdate()

	switch {
	case err != nil:
		log.Printf("Auto-update failed: %v", err)

		return
	case !isUpdateFound:
		log.Println("We are running the latest version, OK!")

		return
	}

	log.Printf("Newer version of the application is found [version=%s]", newVersion)
	log.Printf("What's new:\n%s", changeLog)

	if !doRestartOnUpdate {
		log.Println("Auto restart is disabled, restart the application manually to apply changes!")

		return
	}

	log.Println("Auto restart is enabled, restarting the application to run a new version")

	if err = restart("-skip-update-check-on-start"); err != nil {
		log.Printf("Failed to restart the application after the update to the new version: %v", err)
		log.Println("Restart the application manually to apply changes!")
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
