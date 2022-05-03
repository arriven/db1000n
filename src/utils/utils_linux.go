package utils

import (
	"syscall"

	sys "golang.org/x/sys/unix"
)

func UpdateRLimit() error {
	var rLimit sys.Rlimit

	err := sys.Getrlimit(sys.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err
	}

	rLimit.Cur = rLimit.Max

	return sys.Setrlimit(sys.RLIMIT_NOFILE, &rLimit)
}

func BindToInterface(name string) func(network, address string, conn syscall.RawConn) error {
	return func(network, address string, conn syscall.RawConn) error {
		var operr error

		if err := conn.Control(func(fd uintptr) {
			operr = syscall.BindToDevice(int(fd), name)
		}); err != nil {
			return err
		}

		return operr
	}
}
