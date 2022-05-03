package utils

import "syscall"

func UpdateRLimit() error {
	return nil
}

func BindToInterface(name string) func(network, address string, conn syscall.RawConn) error {
	return func(network, address string, conn syscall.RawConn) error {
		return nil
	}
}
