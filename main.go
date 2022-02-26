package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// JobArgs comment for linter
type JobArgs = json.RawMessage

type job = func(JobArgs) error

// JobConfig comment for linter
type JobConfig struct {
	Type string
	Args JobArgs
}

var jobs = map[string]job{
	"http": httpJob,
}

// Config comment for linter
type Config struct {
	Jobs []JobConfig
}

func httpJob(args JobArgs) error {
	type httpJobConfig struct {
		Path       string
		Method     string
		Body       json.RawMessage
		IntervalMs int `json:"interval_ms,omitempty"`
		Count      int `json:"count,omitempty"`
	}
	// Init defaults
	jobConfig := httpJobConfig{
		Count:      1000,
		IntervalMs: 0,
	}
	// Read from json
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}
	for i := 0; i < jobConfig.Count; i++ {
		switch strings.ToLower(jobConfig.Method) {
		case "get":
			resp, err := http.Get(jobConfig.Path)
			if err != nil {
				log.Printf("error sending get request to [%s]: %v", jobConfig.Path, err)
				continue
			}
			resp.Body.Close() // No need for response
			if resp.StatusCode >= 400 {
				log.Printf("bad response from [%s]: status code %v", jobConfig.Path, resp.StatusCode)
			} else {
				log.Printf("successful http response")
			}
		default:
			log.Printf("method not implemented - %s", jobConfig.Method)
		}
		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}
	return nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "path to a config file, can be web endpoint")
	flag.Parse()
	var configBytes []byte
	var err error
	if configURL, err := url.ParseRequestURI(configPath); err == nil {
		resp, err := http.Get(configURL.String())
		if err != nil {
			fmt.Printf("error sending get request to [%s]: %v\n", configURL.String(), err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			fmt.Println("bad response:", resp.StatusCode)
			return
		}
		configBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response body: %v\n", err)
			return
		}
	} else {
		configBytes, err = os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("error reading file at [%s]: %v\n", configPath, err)
			return
		}
	}
	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Printf("error parsing json config: %v\n", err)
		return
	}
	for _, jobDesc := range config.Jobs {
		if job, ok := jobs[jobDesc.Type]; ok {
			go job(jobDesc.Args)
		} else {
			log.Printf("no such job - %s", jobDesc.Type)
		}
	}
	fmt.Scanln()
}
