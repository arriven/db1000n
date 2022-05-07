package spdnsclient

import (
	"context"
	"errors"
	"math/rand"
)

// Various errors contained in DNSError.
var (
	errNoSuchHost = errors.New("[SP] no such host")

	// For both read and write operations.
	errCanceled = errors.New("[SP] operation was canceled")
)

// errTimeout exists to return the historical "i/o timeout" string
// for context.DeadlineExceeded. See mapErr.
// It is also used when Dialer.Deadline is exceeded.
//
// TODO(iant): We could consider changing this to os.ErrDeadlineExceeded
// in the future, but note that that would conflict with the TODO
// at mapErr that suggests changing it to context.DeadlineExceeded.
var errTimeout error = &timeoutError{}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "[SP] i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func myFastrand() int {
	return rand.Int()
}

func randInt() int {
	x, y := myFastrand(), myFastrand() // 32-bit halves
	u := uint(x)<<31 ^ uint(int32(y))  // full uint, even on 64-bit systems; avoid 32-bit shift on 32-bit systems
	i := int(u >> 1)                   // clear sign bit, even on 32-bit systems
	return i
}

// mapErr maps from the context errors to the historical internal net
// error values.
//
// TODO(bradfitz): get rid of this after adjusting tests and making
// context.DeadlineExceeded implement net.Error?
func mapErr(err error) error {
	switch err {
	case context.Canceled:
		return errCanceled
	case context.DeadlineExceeded:
		return errTimeout
	default:
		return err
	}
}
