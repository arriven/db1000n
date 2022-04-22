package config

import (
	"bytes"
	"flag"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
)

// NewUpdaterOptionsWithFlags returns updater options initialized with command line flags.
func NewUpdaterOptionsWithFlags() (updaterMode *bool, destinationPath *string) {
	return flag.Bool("updater-mode", utils.GetEnvBoolDefault("UPDATER_MODE", false), "Only run config updater"),
		flag.String("updater-destination-config", utils.GetEnvStringDefault("UPDATER_DESTINATION_CONFIG", "config/config.json"),
			"Destination config file to write (only applies if updater-mode is enabled")
}

func UpdateLocal(logger *zap.Logger, destinationPath string, configPaths []string, backupConfig []byte) {
	lastKnownConfig := &RawMultiConfig{Body: backupConfig}

	for {
		if rawConfig := FetchRawMultiConfig(logger, configPaths, lastKnownConfig); !bytes.Equal(lastKnownConfig.Body, rawConfig.Body) {
			if err := writeConfig(logger, rawConfig.Body, destinationPath); err != nil {
				logger.Error("error writing config", zap.Error(err))

				return
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

func writeConfig(logger *zap.Logger, body []byte, destinationPath string) error {
	file, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	defer file.Close()

	size, err := file.Write(body)
	if err != nil {
		return err
	}

	logger.Info("Saved file", zap.String("destination", destinationPath), zap.Int("size", size))

	return nil
}
