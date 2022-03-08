package jobs

import (
	"context"
)

// Args comment for linter
type Args = map[string]interface{}

// Job comment for linter
type Job = func(ctx context.Context, args Args, debug bool) error

// Config comment for linter
type Config struct {
	Type   string
	Count  int
	Filter string
	Args   Args
}

// Get job by type name
func Get(t string) (Job, bool) {
	res, ok := map[string]Job{
		"http":       fasthttpJob,
		"fasthttp":   fasthttpJob,
		"tcp":        tcpJob,
		"udp":        udpJob,
		"slow-loris": slowLorisJob,
		"packetgen":  packetgenJob,
		"dns-blast":  dnsBlastJob,
	}[t]

	return res, ok
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

	if c.Count <= 0 {
		return true
	}

	c.iter++

	return c.iter <= c.Count
}
