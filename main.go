package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type JobArgs = json.RawMessage

type job interface {
	Run(JobArgs)
}

type JobConfig struct {
	Name string
	Args JobArgs
}

type Config struct {
	Jobs []JobConfig
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "f", "", "path to a config file, can be web endpoint")
	flag.Parse()
	var configBytes []byte
	var err error
	if configUrl, err := url.ParseRequestURI(configPath); err == nil {
		resp, err := http.Get(configUrl.String())
		if err != nil {
			fmt.Printf("error getting url at [%s]: %v\n", configUrl.String(), err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusAccepted {
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
	for _, job := range config.Jobs {
		fmt.Println("read job", job.Name, "with args", job.Args)
	}
}
