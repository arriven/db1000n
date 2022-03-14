package updater

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/Arriven/db1000n/src/runner/config"
)

func Run(configPaths []string) {
	configFetcher := config.NewFetcher([]byte{})

	for {
		config := configFetcher.UpdateWithoutUnmarshal(configPaths, "json")

		if config != nil {
			file, err := os.Create("config/config.json")
			if err != nil {
				log.Println("Unable to create config/config.json")

				return
			}

			size, err := io.WriteString(file, string(config.Body))
			if err != nil {
				log.Println("Error while writing to config/config.json")

				return
			}

			defer file.Close()

			configFetcher.LastKnownConfig = *config
			log.Printf("Saved config/config.json with size %d", size)
		}

		time.Sleep(1 * time.Minute)
	}
}
