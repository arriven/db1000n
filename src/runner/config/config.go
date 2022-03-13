// Package config [used for configuring the package]
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/utils"
)

// Config for all jobs to run
type Config struct {
	Jobs []jobs.Config
}

type RawConfig struct {
	body         []byte
	lastModified string
	etag         string
}

type Fetcher struct {
	backupConfig    []byte
	lastKnownConfig RawConfig
}

func NewFetcher(backupConfig []byte) *Fetcher {
	return &Fetcher{
		backupConfig: backupConfig,
		lastKnownConfig: RawConfig{
			body:         nil,
			lastModified: "",
			etag:         "",
		},
	}
}

// fetch tries to read a config from the list of mirrors until it succeeds
func (f *Fetcher) fetch(paths []string) *RawConfig {
	for i := range paths {
		config, err := f.fetchSingle(paths[i])
		if err != nil {
			log.Printf("Failed to fetch config from %q: %v", paths[i], err)

			continue
		}

		log.Printf("Loading config from %q", paths[i])

		return config
	}

	if f.lastKnownConfig.body != nil {
		log.Println("Could not load new config, proceeding with the last known good one")

		return &f.lastKnownConfig
	}

	log.Println("Could not load new config, proceeding with the backup one")

	return &RawConfig{body: f.backupConfig, lastModified: "", etag: ""}
}

// fetchSingle reads a config from a single source
func (f *Fetcher) fetchSingle(path string) (*RawConfig, error) {
	configURL, err := url.ParseRequestURI(path)
	if err != nil {
		res, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return &RawConfig{body: res, lastModified: "", etag: ""}, nil
	}

	const requestTimeout = 20 * time.Second

	client := http.Client{
		Timeout: requestTimeout,
	}
	req, _ := http.NewRequest(http.MethodGet, configURL.String(), nil)
	req.Header.Add("If-None-Match", f.lastKnownConfig.etag)
	req.Header.Add("If-Modified-Since", f.lastKnownConfig.lastModified)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotModified {
		log.Println("Received HTTP 304 Not Modified")

		return &f.lastKnownConfig, nil
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error fetching config, code %d", resp.StatusCode)
	}

	etag := resp.Header.Get("etag")
	lastModified := resp.Header.Get("last-modified")

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &RawConfig{body: res, etag: etag, lastModified: lastModified}, nil
}

// Update the job config from a list of paths or the built-in backup. Returns nil, nil in case of no changes.
func (f *Fetcher) Update(paths []string, format string) *Config {
	newConfig := f.fetch(paths)

	if utils.IsEncrypted(newConfig.body) {
		decryptedConfig, err := utils.Decrypt(newConfig.body)
		if err != nil {
			log.Println("Can't decrypt config")

			return nil
		}

		log.Println("Decrypted config")

		newConfig.body = decryptedConfig
	}

	if bytes.Equal(f.lastKnownConfig.body, newConfig.body) { // Only restart jobs if the new config differs from the current one
		log.Println("The config has not changed. Keep calm and carry on!")

		return nil
	}

	log.Println("New config received, applying")

	var config Config

	switch format {
	case "", "json":
		if err := json.Unmarshal(newConfig.body, &config); err != nil {
			log.Printf("Failed to unmarshal job configs, will keep the current one: %v", err)

			return nil
		}
	case "yaml":
		if err := yaml.Unmarshal(newConfig.body, &config); err != nil {
			log.Printf("Failed to unmarshal job configs, will keep the current one: %v", err)

			return nil
		}
	default:
		log.Printf("Unknown config format: %v", format)
	}

	f.lastKnownConfig = *newConfig

	return &config
}
