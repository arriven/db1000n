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
	Multiplier int
	Limit      int
	Timeout    time.Duration
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
	Count int

	iter int
}

func (c *Counter) Next() bool {
	if c.Count <= 0 {
		return true
	}

	c.iter++

	return c.iter <= c.Count
}
