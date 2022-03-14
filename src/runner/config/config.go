// Package config [used for configuring the package]
package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	Body         []byte
	lastModified string
	etag         string
}

type Fetcher struct {
	backupConfig    []byte
	LastKnownConfig RawConfig
}

func NewFetcher(backupConfig []byte) *Fetcher {
	return &Fetcher{
		backupConfig: backupConfig,
		LastKnownConfig: RawConfig{
			Body:         nil,
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

	if f.LastKnownConfig.Body != nil {
		log.Println("Could not load new config, proceeding with the last known good one")

		return &f.LastKnownConfig
	}

	log.Println("Could not load new config, proceeding with the backup one")

	return &RawConfig{Body: f.backupConfig, lastModified: "", etag: ""}
}

// fetchSingle reads a config from a single source
func (f *Fetcher) fetchSingle(path string) (*RawConfig, error) {
	configURL, err := url.ParseRequestURI(path)
	// absolute paths can be interpreted as a URL with no schema, need to check for that explicitly
	if err != nil || filepath.IsAbs(path) {
		res, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return &RawConfig{Body: res, lastModified: "", etag: ""}, nil
	}

	const requestTimeout = 20 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, configURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("If-None-Match", f.LastKnownConfig.etag)
	req.Header.Add("If-Modified-Since", f.LastKnownConfig.lastModified)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotModified {
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

	return &RawConfig{Body: res, etag: etag, lastModified: lastModified}, nil
}

func (f *Fetcher) UpdateWithoutUnmarshal(paths []string, format string) *RawConfig {
	newConfig := f.fetch(paths)

	if utils.IsEncrypted(newConfig.Body) {
		decryptedConfig, err := utils.Decrypt(newConfig.Body)
		if err != nil {
			log.Println("Can't decrypt config")

			return nil
		}

		log.Println("Decrypted config")

		newConfig.Body = decryptedConfig
	}

	if bytes.Equal(f.LastKnownConfig.Body, newConfig.Body) { // Only restart jobs if the new config differs from the current one
		log.Println("The config has not changed. Keep calm and carry on!")

		return nil
	}

	return newConfig
}

// Update the job config from a list of paths or the built-in backup. Returns nil, nil in case of no changes.
func (f *Fetcher) Update(paths []string, format string) *Config {
	newConfig := f.UpdateWithoutUnmarshal(paths, format)

	if newConfig == nil {
		return nil
	}

	log.Println("New config received, applying")

	var config Config

	switch format {
	case "", "json":
		if err := json.Unmarshal(newConfig.Body, &config); err != nil {
			log.Printf("Failed to unmarshal job configs, will keep the current one: %v", err)

			return nil
		}
	case "yaml":
		if err := yaml.Unmarshal(newConfig.Body, &config); err != nil {
			log.Printf("Failed to unmarshal job configs, will keep the current one: %v", err)

			return nil
		}
	default:
		log.Printf("Unknown config format: %v", format)
	}

	f.LastKnownConfig = *newConfig

	return &config
}
