package utils

import (
	"go.uber.org/zap"
)

func UpdateRLimit(logger *zap.Logger) error {
	return nil
}

func BindToInterface(name string) func(network, address string, conn syscall.RawConn) error {
	return func(network, address string, conn syscall.RawConn) error {
		return nil
	}
}
