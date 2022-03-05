package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Arriven/db1000n/src/jobs"
	"github.com/Arriven/db1000n/src/utils"
)

// Config for all jobs to run
type Config struct {
	Jobs []jobs.Config
}

func FetchConfig(configPath string) (*Config, error) {
	defer utils.PanicHandler()

	var configBytes []byte
	if configURL, err := url.ParseRequestURI(configPath); err == nil {
		resp, err := http.Get(configURL.String())
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("error fetching config, code %d", resp.StatusCode)
		}

		if configBytes, err = io.ReadAll(resp.Body); err != nil {
			return nil, err
		}
	} else if configBytes, err = os.ReadFile(configPath); err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		fmt.Printf("error parsing json config: %v\n", err)
		return nil, err
	}

	return &config, nil
}

func UpdateConfig(configPath, backupConfig string) (config *Config, err error) {
	configPaths := strings.Split(configPath, ",")
	for _, path := range configPaths {
		config, err = FetchConfig(path)
		if err == nil {
			return config, nil
		}
	}
	err = json.Unmarshal([]byte(backupConfig), &config)
	return config, err
}
