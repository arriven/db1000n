package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// JobArgs comment for linter
type JobArgs = json.RawMessage

type job = func(context.Context, JobArgs) error

// JobConfig comment for linter
type JobConfig struct {
	Type  string
	Count int
	Args  JobArgs
}

var jobs = map[string]job{
	"http": httpJob,
	"tcp":  tcpJob,
	"udp":  udpJob,
}

// Config comment for linter
type Config struct {
	Jobs []JobConfig
}

// BasicJobConfig comment for linter
type BasicJobConfig struct {
	IntervalMs int `json:"interval_ms,omitempty"`
	Count      int `json:"count,omitempty"`

	iter int
}

// Next comment for linter
func (c *BasicJobConfig) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}
	if c.Count > 0 {
		defer func() { c.iter++ }()
		return c.iter < c.Count
	}
	return true
}

func randomUUID() string {
	return uuid.New().String()
}

func parseByteTemplate(input []byte) []byte {
	return []byte(parseStringTemplate(string(input)))
}

func parseStringTemplate(input string) string {
	funcMap := template.FuncMap{
		"random_uuid": randomUUID,
	}
	// TODO: consider adding ability to populate custom data
	tmpl, err := template.New("test").Funcs(funcMap).Parse(input)
	if err != nil {
		log.Println(err)
		return input
	}
	var output strings.Builder
	err = tmpl.Execute(&output, nil)
	if err != nil {
		log.Println(err)
		return input
	}

	return output.String()
}

func httpJob(ctx context.Context, args JobArgs) error {
	type httpJobConfig struct {
		BasicJobConfig
		Path    string
		Method  string
		Body    json.RawMessage
		Headers map[string]string
	}
	var jobConfig httpJobConfig
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}
	for jobConfig.Next(ctx) {
		req, err := http.NewRequest(parseStringTemplate(jobConfig.Method), parseStringTemplate(jobConfig.Path), bytes.NewReader(parseByteTemplate(jobConfig.Body)))
		if err != nil {
			log.Printf("error creating request: %v", err)
			continue
		}
		for key, value := range jobConfig.Headers {
			req.Header.Add(parseStringTemplate(key), parseStringTemplate(value))
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("error sending request %v: %v", req, err)
			continue
		}
		resp.Body.Close() // No need for response
		if resp.StatusCode >= 400 {
			log.Printf("bad response from [%s]: status code %v", req.URL, resp.StatusCode)
		} else {
			log.Printf("successful http response")
		}
		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}
	return nil
}

// RawNetJobConfig comment for linter
type RawNetJobConfig struct {
	BasicJobConfig
	Address string
	Body    json.RawMessage
}

func tcpJob(ctx context.Context, args JobArgs) error {
	type tcpJobConfig struct {
		RawNetJobConfig
	}
	var jobConfig tcpJobConfig
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", parseStringTemplate(jobConfig.Address))
	if err != nil {
		return err
	}
	for jobConfig.Next(ctx) {
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			log.Printf("error connecting to [%v]: %v", tcpAddr, err)
			continue
		}

		_, err = conn.Write(parseByteTemplate(jobConfig.Body))
		if err != nil {
			log.Printf("error sending body to [%v]: %v", tcpAddr, err)
		} else {
			log.Printf("sent body to [%v]", tcpAddr)
		}
		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}
	return nil
}

func udpJob(ctx context.Context, args JobArgs) error {
	type udpJobConfig struct {
		RawNetJobConfig
	}
	var jobConfig udpJobConfig
	err := json.Unmarshal(args, &jobConfig)
	if err != nil {
		return err
	}
	udpAddr, err := net.ResolveUDPAddr("udp", parseStringTemplate(jobConfig.Address))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Printf("error connecting to [%v]: %v", udpAddr, err)
		return err
	}

	for jobConfig.Next(ctx) {
		_, err = conn.Write(parseByteTemplate(jobConfig.Body))
		if err != nil {
			log.Printf("error sending body to [%v]: %v", udpAddr, err)
		} else {
			log.Printf("sent body to [%v]", udpAddr)
		}
		time.Sleep(time.Duration(jobConfig.IntervalMs) * time.Millisecond)
	}
	return nil
}

func fetchConfig(configPath string) (*Config, error) {
	var configBytes []byte
	var err error
	if configURL, err := url.ParseRequestURI(configPath); err == nil {
		resp, err := http.Get(configURL.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, err
		}
		configBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		configBytes, err = os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
	}
	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Printf("error parsing json config: %v\n", err)
		return nil, err
	}
	return &config, nil
}

func main() {
	var configPath string
	var refreshTimeout time.Duration
	flag.StringVar(&configPath, "c", "https://raw.githubusercontent.com/db1000n-coordinators/LoadTestConfig/main/config.json", "path to a config file, can be web endpoint")
	flag.DurationVar(&refreshTimeout, "r", time.Minute, "refresh timeout for updating the config")
	flag.Parse()
	var cancel context.CancelFunc
	defer cancel()
	for {
		config, err := fetchConfig(configPath)
		if err != nil {
			fmt.Printf("error fetching json config: %v\n", err)
			continue
		}
		if cancel != nil {
			cancel()
		}
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		for _, jobDesc := range config.Jobs {
			if jobDesc.Count < 1 {
				jobDesc.Count = 1
			}
			if job, ok := jobs[jobDesc.Type]; ok {
				for i := 0; i < jobDesc.Count; i++ {
					go job(ctx, jobDesc.Args)
				}
			} else {
				log.Printf("no such job - %s", jobDesc.Type)
			}
		}
		time.Sleep(refreshTimeout)
	}
}
