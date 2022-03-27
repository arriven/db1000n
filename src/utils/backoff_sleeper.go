package utils

import (
	"context"
	"time"
)

// BackoffSleeper sleeps for the specified interval of time and implements the binary exponential backoff for rate limiting.
//
// Not concurrency-safe.
type BackoffSleeper struct {
	IntervalMs int `mapstructure:"interval_ms,omitempty"`

	backoff bool // backoff mode on/off
	factor  int  // rate limiting factor
}

// Sleep adjusts the backoff rate limiting factor depending on whether the backoff mode is on or off.
// It returns false if interrupted by the context cancellation; true otherwise.
func (s *BackoffSleeper) Sleep(ctx context.Context) bool {
	s.adjustBackoff()

	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Duration(s.factor*s.IntervalMs) * time.Millisecond):
		return true
	}
}

func (s *BackoffSleeper) adjustBackoff() {
	if s.factor < 1 { // Avoid the need in a constructor
		s.factor = 1
	}

	const base = 2

	switch s.backoff {
	case true:
		const limit = 4096

		if s.factor < limit {
			s.factor *= base
		}
	case false:
		if s.factor > 1 {
			s.factor /= base
		}
	}
}

// EnableBackoff starts the exponential rate limiting.
func (s *BackoffSleeper) EnableBackoff() { s.backoff = true }

// DisableBackoff starts the exponential rate recovery.
func (s *BackoffSleeper) DisableBackoff() { s.backoff = false }

// ResetBackoff disables the rate limiting and recovers the rate immediately.
func (s *BackoffSleeper) ResetBackoff() {
	s.backoff = false
	s.factor = 1
}
