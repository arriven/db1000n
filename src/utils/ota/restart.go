package ota

import (
	"log"
	"os"
	"syscall"

	"github.com/pkg/errors"
)

func Restart(extraArgs ...string) error {
	executable, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "failed to resolve the path to the current executable")
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to resolve the current working directory")
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
		args = mergeExtraArgs(os.Args, extraArgs)
	} else {
		args = os.Args
	}

	fork, err := syscall.ForkExec(executable, args, execSpec)
	if err != nil {
		return errors.Wrap(err, "failed to spawn a new process")
	}

	log.Printf("new process has been started successfully [old_pid=%d,new_pid=%d]\n",
		os.Getpid(), fork)

	os.Exit(0)
	return nil
}

func mergeExtraArgs(osArgs, extraArgs []string) []string {
	osArgsC := make([]string, len(osArgs))
	copy(osArgsC, osArgs)

	isPresent := func(what string, where []string) bool {
		isFound := false
		for i := range where {
			if where[i] == what {
				isFound = true
			}
		}
		return isFound
	}

	for i := range extraArgs {
		if !isPresent(extraArgs[i], osArgs) {
			osArgsC = append(osArgsC, extraArgs[i])
		}
	}

	return osArgsC
}
