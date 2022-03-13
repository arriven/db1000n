//go:build linux
// +build linux

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

//nolint:makezero,wsl // Makezero is skipped because it's expected to append arguments, same for WSL
func appendArgIfNotPresent(osArgs, extraArgs []string) []string {
	osArgsC := make([]string, len(osArgs))
	copy(osArgsC, osArgs)

	nothing := struct{}{}
	osArgsSet := make(map[string]struct{})

	for _, osArg := range osArgs {
		osArgsSet[osArg] = nothing
	}

	acceptedExtraArgs := make([]string, 0)
	for _, extraArg := range extraArgs {
		_, isAlreadyOSArg := osArgsSet[extraArg]
		if !isAlreadyOSArg {
			acceptedExtraArgs = append(acceptedExtraArgs, extraArg)
		}
	}

	return append(osArgsC, acceptedExtraArgs...)
}
