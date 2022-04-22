//go:build windows
// +build windows

package ota

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

func restart(logger *zap.Logger, extraArgs ...string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate the executable file: %w", err)
	}

	// A specific way the `cmd.exe` processes the `start` command: it takes the
	// first quoted substring as a process name! (https://superuser.com/a/1656458)
	cmdLine := fmt.Sprintf(`/C start "process" "%s"`, execPath)
	args := []string{}

	if len(extraArgs) != 0 {
		args = appendArgIfNotPresent(os.Args[1:], extraArgs)
	} else {
		args = os.Args[1:]
	}

	if len(args) > 0 {
		cmdLine += " " + strings.Join(args, " ")
	}

	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: cmdLine}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start the command: %w", err)
	}

	os.Exit(0)

	return nil
}
