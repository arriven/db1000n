//go:build windows
// +build windows

package packetgen

import "errors"

// unsupported on windows
type rawConn struct{}

func openRawConn() (*rawConn, error) {
	return nil, errors.New("raw connections not supported on windows")
}

func (conn *rawConn) Write(packet Packet) (n int, err error) {
	return 0, errors.New("raw connections not supported on windows")
}

func (conn *rawConn) Close() error {
	return nil
}

func (conn *rawConn) Target() string { return "raw://" }

func (conn *rawConn) Read(_ []byte) (int, error) { return 0, nil }
