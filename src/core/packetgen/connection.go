// MIT License

// Copyright (c) [2022] [Bohdan Ivashko (https://github.com/Arriven)]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package packetgen

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"golang.org/x/net/ipv6"

	"github.com/Arriven/db1000n/src/utils"
)

// ConnectionConfig describes which network to use when sending packets
type ConnectionConfig struct {
	Type string
	Args map[string]interface{}
}

func OpenConnection(c ConnectionConfig) (Connection, error) {
	switch c.Type {
	case "raw":
		var cfg rawConnectionConfig
		if err := utils.Decode(c.Args, &cfg); err != nil {
			return nil, fmt.Errorf("error decoding connection config: %w", err)
		}

		return openRawConnection(cfg)
	default:
		return nil, fmt.Errorf("unknown connection type: %v", c.Type)
	}
}

type Connection interface {
	Write(Packet) (int, error)
}

// raw ipv4/ipv6 connection
type rawConnectionConfig struct {
	Name    string
	Address string
}

type rawConn struct {
	*ipv6.PacketConn
}

// openRawConnection opens a raw ip network connection based on the provided config
// use ipv6 as it also supports ipv4
func openRawConnection(c rawConnectionConfig) (*rawConn, error) {
	packetConn, err := net.ListenPacket(c.Name, c.Address)
	if err != nil {
		return nil, err
	}

	return &rawConn{PacketConn: ipv6.NewPacketConn(packetConn)}, nil
}

func (conn rawConn) Write(packet Packet) (n int, err error) {
	payloadBuf := gopacket.NewSerializeBuffer()

	if err = packet.Serialize(payloadBuf); err != nil {
		return 0, fmt.Errorf("error serializing packet: %w", err)
	}

	return conn.WriteTo(payloadBuf.Bytes(), nil, &net.IPAddr{IP: packet.IP()})
}
