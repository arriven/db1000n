//go:build windows || netbsd || solaris || aix || dragonfly || freebsd || (illumos && !linux && !darwin)
// +build windows netbsd solaris aix dragonfly freebsd illumos,!linux,!darwin

package ota

import "errors"

func Restart(extraArgs ...string) error {
	return errors.New("restart on the Windows system is not available")
}
