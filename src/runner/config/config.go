package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Arriven/db1000n/src/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/Arriven/db1000n/src/jobs"
	"gopkg.in/yaml.v3"
)

// Config for all jobs to run
type Config struct {
	Jobs []jobs.Config
}

// Fetch tries to read a config from the list of mirrors until it succeeds
func Fetch(paths []string) ([]byte, error) {
	for i := range paths {
		res, err := FetchSingle(paths[i])
		if err != nil {
			log.Printf("Failed to fetch config from %q: %v", paths[i], err)
			continue
		}

		log.Printf("Loading config from %q", paths[i])

		return res, nil
	}

	return nil, errors.New("config fetch failed")
}

// FetchSingle reads a config from a single source
func FetchSingle(path string) ([]byte, error) {
	configURL, err := url.ParseRequestURI(path)
	if err != nil {
		res, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	resp, err := http.Get(configURL.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error fetching config, code %d", resp.StatusCode)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func Update(paths []string, current, backup []byte, format string) (*Config, []byte) {
	newRawConfig, err := Fetch(paths)
	if err != nil {
		if current != nil {
			log.Println("Could not load new config, proceeding with last known good config")
			newRawConfig = current
		} else {
			log.Println("Could not load new config, proceeding with backupConfig")
			newRawConfig = backup
		}
	}

	if !bytes.Equal(current, newRawConfig) { // Only restart jobs if the new config differs from the current one
		log.Println("New config received, applying")
		if utils.IsEncrypted(newRawConfig) {
			decryptedConfig, err := utils.DecryptConfig(newRawConfig)
			if err != nil {
				log.Println("Can't decrypt config")
				return nil, nil
			}
			log.Println("Decrypted config")
			newRawConfig = decryptedConfig
		}
		var config Config
		if format == "" {
			format = "json"
		}
		switch format {
		case "json":
			if err := json.Unmarshal(newRawConfig, &config); err != nil {
				log.Printf("Failed to unmarshal job configs: %v", err)
				return nil, nil
			}
		case "yaml":
			if err := yaml.Unmarshal(newRawConfig, &config); err != nil {
				log.Printf("Failed to unmarshal job configs: %v", err)
				return nil, nil
			}
		default:
			log.Printf("Unknown config format: %v", format)
		}
		return &config, newRawConfig
	}
	return nil, nil
}
