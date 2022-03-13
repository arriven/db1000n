//+build windows

package ota

import "errors"

func Restart(extraArgs ...string) error {
	return errors.New("restart on the Windows system is not available")
}
