package packetgen

import (
	"net"

	"golang.org/x/net/ipv6"
)

// ConnectionConfig describes which network to use when sending packets
type ConnectionConfig struct {
	Name    string
	Address string
}

// OpenRawConnection opens a raw ip network connection based on the provided config
// use ipv6 as it also supports ipv4
func OpenRawConnection(c ConnectionConfig) (*ipv6.PacketConn, error) {
	packetConn, err := net.ListenPacket(c.Name, c.Address)
	if err != nil {
		return nil, err
	}

	return ipv6.NewPacketConn(packetConn), nil
}
