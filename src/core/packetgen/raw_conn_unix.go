//go:build !windows
// +build !windows

package packetgen

import (
	"fmt"
	"syscall"

	"github.com/google/gopacket"
)

type rawConn struct {
	fd  int
	buf gopacket.SerializeBuffer
}

// openRawConn opens a raw ip network connection based on the provided config
// use ipv6 as it also supports ipv4
func openRawConn() (*rawConn, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		return nil, err
	}

	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		return nil, err
	}

	return &rawConn{
		fd:  fd,
		buf: gopacket.NewSerializeBuffer(),
	}, nil
}

func (conn *rawConn) Write(packet Packet) (n int, err error) {
	if err := packet.Serialize(conn.buf); err != nil {
		return 0, fmt.Errorf("error serializing packet: %w", err)
	}

	addr := &syscall.SockaddrInet4{}

	// ipv6 is not supported for now
	copy(addr.Addr[:], packet.IP().To4())

	return 0, syscall.Sendto(conn.fd, conn.buf.Bytes(), 0, addr)
}

func (conn *rawConn) Close() error {
	return syscall.Close(conn.fd)
}

func (conn *rawConn) Target() string { return "raw://" }

func (conn *rawConn) Read(_ []byte) (int, error) { return 0, nil }
