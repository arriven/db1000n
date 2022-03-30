package utils

import (
	"context"
	"time"
)

func Sleep(ctx context.Context, t time.Duration) bool {
	select {
	case <-time.After(t):
		return true
	case <-ctx.Done():
		return false
	}
}

type BackoffConfig struct {
	Multiplier int           `mapstructure:"backoff_multiplier"`
	Limit      int           `mapstructure:"backoff_limit"`
	Timeout    time.Duration `mapstructure:"backoff_timeout"`
}

func DefaultBackoffConfig() BackoffConfig {
	const (
		defaultMultiplier = 10
		defaultLimit      = 6
	)

	return BackoffConfig{Multiplier: defaultMultiplier, Limit: defaultLimit, Timeout: time.Microsecond}
}

func NonNilBackoffConfigOrDefault(c *BackoffConfig, defaultConfig BackoffConfig) *BackoffConfig {
	if c != nil {
		return c
	}

	return &defaultConfig
}

type BackoffController struct {
	BackoffConfig
	count int
}

func NewBackoffController(c *BackoffConfig) BackoffController {
	return BackoffController{BackoffConfig: *NonNilBackoffConfigOrDefault(c, DefaultBackoffConfig())}
}

func (c BackoffController) GetTimeout() time.Duration {
	result := c.Timeout
	for i := 0; i < c.count; i++ {
		result *= time.Duration(c.Multiplier)
	}

	return result
}

func (c *BackoffController) Increment() *BackoffController {
	if c.count < c.Limit {
		c.count++
	}

	// return pointer to itself for usability
	return c
}

func (c *BackoffController) Reset() {
	c.count = 0
}

type Counter struct {
	Count int `mapstructure:"count,omitempty"`

	iter int
}

func (c *Counter) Next() bool {
	if c.Count <= 0 {
		return true
	}

	c.iter++

	return c.iter <= c.Count
}
