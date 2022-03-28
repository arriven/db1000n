package config

import (
	"bytes"
	"flag"
	"log"
	"os"
	"time"

	"github.com/Arriven/db1000n/src/utils"
)

// NewUpdaterOptionsWithFlags returns updater options initialized with command line flags.
func NewUpdaterOptionsWithFlags() (updaterMode *bool, destinationPath *string) {
	return flag.Bool("updater-mode", utils.GetEnvBoolDefault("UPDATER_MODE", false), "Only run config updater"),
		flag.String("updater-destination-config", utils.GetEnvStringDefault("UPDATER_DESTINATION_CONFIG", "config/config.json"),
			"Destination config file to write (only applies if updater-mode is enabled")
}

func UpdateLocal(destinationPath string, configPaths []string, backupConfig []byte) {
	lastKnownConfig := &RawMultiConfig{Body: backupConfig}

	for {
		if rawConfig := FetchRawMultiConfig(configPaths, lastKnownConfig); !bytes.Equal(lastKnownConfig.Body, rawConfig.Body) {
			if err := writeConfig(rawConfig.Body, destinationPath); err != nil {
				log.Printf("Error writing config: %v", err)

				return
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

func writeConfig(body []byte, destinationPath string) error {
	file, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	defer file.Close()

	size, err := file.Write(body)
	if err != nil {
		return err
	}

	log.Printf("Saved %s with size %d", destinationPath, size)

	return nil
}
