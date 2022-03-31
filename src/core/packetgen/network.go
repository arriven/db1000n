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
	"github.com/google/gopacket/layers"

	"github.com/Arriven/db1000n/src/utils"
)

func BuildNetworkLayer(c LayerConfig) (gopacket.NetworkLayer, error) {
	switch c.Type {
	case "":
		return nil, nil
	case "ipv4":
		var packetConfig IPPacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return buildIPV4Packet(packetConfig), nil
	case "ipv6":
		var packetConfig IPPacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return buildIPV6Packet(packetConfig), nil
	default:
		return nil, fmt.Errorf("unsupported network layer type %s", c.Type)
	}
}

// IPPacketConfig describes ip layer configuration
type IPPacketConfig struct {
	SrcIP        string
	DstIP        string
	NextProtocol *int
}

// buildIPV4Packet generates a layers.IPv4 and returns it with source IP address and destination IP address
func buildIPV4Packet(c IPPacketConfig) *layers.IPv4 {
	const ipv4 = 4

	next := layers.IPProtocolTCP
	if c.NextProtocol != nil {
		next = layers.IPProtocol(*c.NextProtocol)
	}

	return &layers.IPv4{
		SrcIP:    net.ParseIP(c.SrcIP).To4(),
		DstIP:    net.ParseIP(c.DstIP).To4(),
		Version:  ipv4,
		Protocol: next,
	}
}

// buildIPV6Packet generates a layers.IPv6 and returns it with source IP address and destination IP address
func buildIPV6Packet(c IPPacketConfig) *layers.IPv6 {
	const ipv6 = 6

	next := layers.IPProtocolTCP
	if c.NextProtocol != nil {
		next = layers.IPProtocol(*c.NextProtocol)
	}

	return &layers.IPv6{
		SrcIP:      net.ParseIP(c.SrcIP).To16(),
		DstIP:      net.ParseIP(c.DstIP).To16(),
		Version:    ipv6,
		NextHeader: next,
	}
}
