package updater

import (
	"bytes"
	"flag"
	"log"
	"os"
	"time"

	"github.com/Arriven/db1000n/src/runner/config"
	"github.com/Arriven/db1000n/src/utils"
)

// NewOptionsWithFlags returns updater options initialized with command line flags.
func NewOptionsWithFlags() (updaterMode *bool, destinationConfig *string) {
	return flag.Bool("updater-mode", utils.GetEnvBoolDefault("UPDATER_MODE", false), "Only run config updater"),
		flag.String("updater-destination-config", utils.GetEnvStringDefault("UPDATER_DESTINATION_CONFIG", "config/config.json"),
			"Destination config file to write (only applies if updater-mode is enabled")
}

func Run(destinationConfig string, configPaths []string, backupConfig []byte) {
	lastKnownConfig := &config.RawConfig{Body: backupConfig}

	for {
		rawConfig := config.FetchRawConfig(configPaths, lastKnownConfig)

		if !bytes.Equal(lastKnownConfig.Body, rawConfig.Body) {
			err := writeConfig(rawConfig.Body, destinationConfig)
			if err != nil {
				log.Printf("Error writing config: %v", err)

				return
			}

			lastKnownConfig = rawConfig
		}

		time.Sleep(1 * time.Minute)
	}
}

func writeConfig(body []byte, destinationConfig string) error {
	file, err := os.Create(destinationConfig)
	if err != nil {
		return err
	}

	defer file.Close()

	size, err := file.Write(body)
	if err != nil {
		return err
	}

	log.Printf("Saved %s with size %d", destinationConfig, size)

	return nil
}
