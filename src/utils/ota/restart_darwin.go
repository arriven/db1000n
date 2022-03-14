//go:build darwin
// +build darwin

package ota

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

func Restart(extraArgs ...string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve the path to the current executable: %w", err)
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to resolve the current working directory: %w", err)
	}

	execSpec := &syscall.ProcAttr{
		Dir: workingDirectory,
		Env: os.Environ(),
		Files: []uintptr{
			os.Stdin.Fd(),
			os.Stdout.Fd(),
			os.Stderr.Fd(),
		},
	}

	var args []string

	if len(extraArgs) != 0 {
		args = appendArgIfNotPresent(os.Args[1:], extraArgs)
	} else {
		args = os.Args[1:]
	}

	fork, err := syscall.ForkExec(executable, args, execSpec)
	if err != nil {
		return fmt.Errorf("failed to spawn a new process: %w", err)
	}

	log.Printf("new process has been started successfully [old_pid=%d,new_pid=%d]\n",
		os.Getpid(), fork)

	os.Exit(0)

	return nil
}
