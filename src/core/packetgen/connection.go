package packetgen

import (
	"net"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// ConnectionConfig describes which network to use when sending packets
type ConnectionConfig struct {
	Name    string
	Address string
}

// OpenRawConnectionV4 opens a raw ip network connection based on the provided config
func OpenRawConnectionV4(c ConnectionConfig) (*ipv4.RawConn, error) {
	packetConn, err := net.ListenPacket(c.Name, c.Address)
	if err != nil {
		return nil, err
	}

	return ipv4.NewRawConn(packetConn)
}

// OpenRawConnection opens a raw ip network connection based on the provided config
func OpenRawConnectionV6(c ConnectionConfig) (*ipv6.PacketConn, error) {
	packetConn, err := net.ListenPacket(c.Name, c.Address)
	if err != nil {
		return nil, err
	}

	return ipv6.NewPacketConn(packetConn), nil
}
