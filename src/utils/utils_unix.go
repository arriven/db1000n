//go:build !windows
// +build !windows

package utils

import (
	"go.uber.org/zap"
	sys "golang.org/x/sys/unix"
)

func UpdateRLimit(logger *zap.Logger) error {
	var rLimit sys.Rlimit

	err := sys.Getrlimit(sys.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	rLimit.Cur = rLimit.Max

	return sys.Setrlimit(sys.RLIMIT_NOFILE, &rLimit)
}
