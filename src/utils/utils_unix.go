//go:build !windows
// +build !windows

package utils

import (
	"net"
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

func getSockaddrByName(name string) syscall.Sockaddr {
	ief, err := net.InterfaceByName(name)
	if err != nil {
		return nil
	}

	addrs, err := ief.Addrs()
	if err != nil {
		return nil
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		if ipBytes := ipNet.IP.To4(); ipBytes != nil {
			var sa4 syscall.SockaddrInet4

			copy(sa4.Addr[:], ipBytes)

			return &sa4
		} else if ipBytes := ipNet.IP.To16(); ipBytes != nil {
			var sa16 syscall.SockaddrInet6

			copy(sa16.Addr[:], ipBytes)
			sa16.ZoneId = uint32(ief.Index)

			return &sa16
		}
	}

	return nil
}

func BindToInterface(name string) func(network, address string, conn syscall.RawConn) error {
	return func(network, address string, conn syscall.RawConn) error {
		sockAddr := getSockaddrByName(name)
		if sockAddr == nil {
			return nil
		}

		var operr error

		if err := conn.Control(func(fd uintptr) {
			operr = syscall.Bind(int(fd), sockAddr)
		}); err != nil {
			return err
		}

		return operr
	}
}
