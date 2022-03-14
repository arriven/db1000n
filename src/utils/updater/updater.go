package updater

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	"github.com/Arriven/db1000n/src/runner/config"
)

func Run(configPaths []string, backupConfig []byte) {
	lastKnownConfig := &config.RawConfig{Body: backupConfig}

	for {
		rawConfig := config.FetchRawConfig(configPaths, lastKnownConfig)

		if !bytes.Equal(lastKnownConfig.Body, rawConfig.Body) {
			file, err := os.Create("config/config.json")
			if err != nil {
				log.Println("Unable to create config/config.json")

				return
			}

			size, err := io.WriteString(file, string(rawConfig.Body))
			if err != nil {
				log.Println("Error while writing to config/config.json")

				return
			}

			defer file.Close()

			lastKnownConfig = rawConfig

			log.Printf("Saved config/config.json with size %d", size)
		}

		time.Sleep(1 * time.Minute)
	}
}
