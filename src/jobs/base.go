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
	ProxyURL string
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
func Get(t string) (Job, bool) {
	res, ok := map[string]Job{
		"http":         fastHTTPJob,
		"http-flood":   fastHTTPJob,
		"http-request": singleRequestJob,
		"tcp":          tcpJob,
		"udp":          udpJob,
		"slow-loris":   slowLorisJob,
		"packetgen":    packetgenJob,
		"dns-blast":    dnsBlastJob,
		"sequence":     sequenceJob,
		"parallel":     parallelJob,
		"log":          logJob,
		"set-value":    setVarJob,
		"check":        checkJob,
	}[t]

	return res, ok
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
