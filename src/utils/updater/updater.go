package updater

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"github.com/Arriven/db1000n/src/runner/config"
)

func Run(destinationConfig string, configPaths []string, backupConfig []byte) {
	lastKnownConfig := &config.RawConfig{Body: backupConfig}

	for {
		rawConfig := config.FetchRawConfig(configPaths, lastKnownConfig)

		if !bytes.Equal(lastKnownConfig.Body, rawConfig.Body) {
			file, err := os.Create(destinationConfig)
			if err != nil {
				log.Printf("Unable to create %s", destinationConfig)

				return
			}

			size, err := io.WriteString(file, string(rawConfig.Body))
			if err != nil {
				log.Printf("Error while writing to %s", destinationConfig)

				return
			}

			defer file.Close()

			lastKnownConfig = rawConfig

			log.Printf("Saved %s with size %d", destinationConfig, size)
		}

		time.Sleep(1 * time.Minute)
	}
}
