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

type BackoffController struct {
	BackoffConfig
	count int
}

func NewBackoffController(c *BackoffConfig) BackoffController {
	if c != nil {
		return BackoffController{BackoffConfig: *c}
	}

	return BackoffController{BackoffConfig: DefaultBackoffConfig()}
}

func (c BackoffController) getTimeout() time.Duration {
	result := c.Timeout
	for i := 0; i < c.count; i++ {
		result *= time.Duration(c.Multiplier)
	}

	return result
}

func (c *BackoffController) increment() {
	if c.count < c.Limit {
		c.count++
	}
}

func (c *BackoffController) reset() {
	c.count = 0
}

func (c *BackoffController) Handle(ctx context.Context, err error) {
	if err == nil {
		c.reset()

		return
	}

	c.increment()
	Sleep(ctx, c.getTimeout())
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
