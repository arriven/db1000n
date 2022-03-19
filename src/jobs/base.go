// Package jobs [contains all the attack types db1000n can simulate]
package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/Arriven/db1000n/src/utils/templates"
)

// Args comment for linter
type Args = map[string]interface{}

// GlobalConfig is a struct meant to pass commandline arguments to every job
type GlobalConfig struct {
	ProxyURL            string
	ScaleFactor         int
	SkipEncrypted       bool
	Debug               bool
	ClientID            string
	EnablePrimitiveJobs bool
}

const (
	isEncryptedContextKey = "is_in_encrypted_context"
)

func isInEncryptedContext(ctx context.Context) bool {
	return ctx.Value(templates.ContextKey(isEncryptedContextKey)) != nil
}

// Job comment for linter
type Job = func(ctx context.Context, logger *zap.Logger, globalConfig GlobalConfig, args Args) (data interface{}, err error)

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
	case "encrypted":
		return encryptedJob
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
