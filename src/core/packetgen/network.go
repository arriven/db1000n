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
	SrcIP        string `mapstructure:"src_ip"`
	DstIP        string `mapstructure:"dst_ip"`
	NextProtocol *int   `mapstructure:"next"`
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
