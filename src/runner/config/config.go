// Package config [used for configuring the package]
package config

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

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

// fetch tries to read a config from the list of mirrors until it succeeds
func fetch(paths []string, lastKnownConfig *RawConfig) *RawConfig {
	for i := range paths {
		config, err := fetchSingle(paths[i], lastKnownConfig)
		if err != nil {
			log.Printf("Failed to fetch config from %q: %v", paths[i], err)

			continue
		}

		log.Printf("Loading config from %q", paths[i])

		return config
	}

	return lastKnownConfig
}

// fetchSingle reads a config from a single source
func fetchSingle(path string, lastKnownConfig *RawConfig) (*RawConfig, error) {
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

	req.Header.Add("If-None-Match", lastKnownConfig.etag)
	req.Header.Add("If-Modified-Since", lastKnownConfig.lastModified)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return lastKnownConfig, nil
	}

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

func FetchRawConfig(paths []string, lastKnownConfig *RawConfig) *RawConfig {
	newConfig := fetch(paths, lastKnownConfig)

	if utils.IsEncrypted(newConfig.Body) {
		decryptedConfig, err := utils.Decrypt(newConfig.Body)
		if err != nil {
			log.Println("Can't decrypt config")

			return lastKnownConfig
		}

		log.Println("Decrypted config")

		newConfig.Body = decryptedConfig
	}

	return newConfig
}

// Update the job config from a list of paths or the built-in backup. Returns nil, nil in case of no changes.
func Unmarshal(body []byte, format string) *Config {
	if body == nil {
		return nil
	}

	log.Println("New config received, applying")

	var config Config

	if err := utils.Unmarshal(body, &config, format); err != nil {
		log.Printf("Failed to unmarshal job configs, will keep the current one: %v", err)

		return nil
	}

	return &config
}
