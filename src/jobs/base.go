// Package jobs [contains all the attack types db1000n can simulate]
package jobs

import (
	"context"
	"time"
)

// Args comment for linter
type Args = map[string]interface{}

// GlobalConfig is a struct meant to pass commandline arguments to every job
type GlobalConfig struct {
	ProxyURL     string
	ProxyListURL string
	ScaleFactor  int
}

// Job comment for linter
type Job = func(ctx context.Context, globalConfig GlobalConfig, args Args, debug bool) (data interface{}, err error)

// Config comment for linter
type Config struct {
	Name   string `mapstructure:"name"`
	Type   string `mapstructure:"type"`
	Count  int    `mapstructure:"count"`
	Filter string `mapstructure:"filter"`
	Args   Args   `mapstructure:"args"`
}

// Get job by type name
func Get(t string) Job {
	switch t {
	case "http", "http-flood":
		return fastHTTPJob
	case "http-request":
		return singleRequestJob
	case "tcp":
		return tcpJob
	case "udp":
		return udpJob
	case "slow-loris":
		return slowLorisJob
	case "packetgen":
		return packetgenJob
	case "dns-blast":
		return dnsBlastJob
	case "sequence":
		return sequenceJob
	case "parallel":
		return parallelJob
	case "log":
		return logJob
	case "set-value":
		return setVarJob
	case "check":
		return checkJob
	case "loop":
		return loopJob
	default:
		return nil
	}
}

// BasicJobConfig comment for linter
type BasicJobConfig struct {
	IntervalMs int `mapstructure:"interval_ms,omitempty"`
	Count      int `mapstructure:"count,omitempty"`

	iter int
}

// Next comment for linter
func (c *BasicJobConfig) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Duration(c.IntervalMs) * time.Millisecond):
		if c.Count <= 0 {
			return true
		}

		c.iter++

		return c.iter <= c.Count
	}
}
