// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package config [used for configuring the package]
package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils"
)

// Args is a generic arguments map.
type Args = map[string]any

// Config for a single job.
type Config struct {
	Name   string
	Type   string
	Count  int
	Filter string
	Args   Args
}

// MultiConfig for all jobs.
type MultiConfig struct {
	Jobs []Config
}

type RawMultiConfig struct {
	Body         []byte
	Protected    bool
	lastModified string
	etag         string
}

// fetch tries to read a config from the list of mirrors until it succeeds
func fetch(logger *zap.Logger, paths []string, lastKnownConfig *RawMultiConfig, skipEncrypted bool) *RawMultiConfig {
	for i := range paths {
		config, err := fetchAndDecrypt(logger, paths[i], lastKnownConfig, skipEncrypted)
		if err != nil {
			continue
		}

		logger.Info("loading config", zap.String("path", paths[i]))

		return config
	}

	return lastKnownConfig
}

func fetchAndDecrypt(logger *zap.Logger, path string, lastKnownConfig *RawMultiConfig, skipEncrypted bool) (*RawMultiConfig, error) {
	config, err := fetchSingle(path, lastKnownConfig)
	if err != nil {
		logger.Warn("failed to fetch config", zap.String("path", path), zap.Error(err))

		return nil, err
	}

	if utils.IsEncrypted(config.Body) {
		if skipEncrypted {
			logger.Warn("can't decrypt config", zap.String("error", "encryption disabled"))

			return nil, fmt.Errorf("encryption disabled")
		}

		decryptedConfig, protected, err := utils.Decrypt(config.Body)
		if err != nil {
			logger.Warn("can't decrypt config", zap.Error(err))

			return nil, err
		}

		logger.Info("decrypted config")

		config.Body = decryptedConfig
		config.Protected = protected
	}

	return config, nil
}

// fetchSingle reads a config from a single source
func fetchSingle(path string, lastKnownConfig *RawMultiConfig) (*RawMultiConfig, error) {
	configURL, err := url.ParseRequestURI(path)
	// absolute paths can be interpreted as a URL with no schema, need to check for that explicitly
	if err != nil || filepath.IsAbs(path) {
		res, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return &RawMultiConfig{Body: res, lastModified: "", etag: ""}, nil
	}

	return fetchURL(configURL, lastKnownConfig)
}

func fetchURL(configURL *url.URL, lastKnownConfig *RawMultiConfig) (*RawMultiConfig, error) {
	const requestTimeout = 20 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, configURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if lastKnownConfig.etag != "" {
		req.Header.Add("If-None-Match", lastKnownConfig.etag)
	}

	if lastKnownConfig.lastModified != "" {
		req.Header.Add("If-Modified-Since", lastKnownConfig.lastModified)
	}

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

	return &RawMultiConfig{Body: res, etag: etag, lastModified: lastModified}, nil
}

// FetchRawMultiConfig retrieves the current config using a list of paths. Falls back to the last known config in case of errors.
func FetchRawMultiConfig(logger *zap.Logger, paths []string, lastKnownConfig *RawMultiConfig, skipEncrypted bool) *RawMultiConfig {
	return fetch(logger, paths, lastKnownConfig, skipEncrypted)
}

// Unmarshal config encoded with the given format.
func Unmarshal(body []byte, format string) *MultiConfig {
	if body == nil {
		return nil
	}

	var config MultiConfig

	if err := utils.Unmarshal(body, &config, format); err != nil {
		return nil
	}

	return &config
}
