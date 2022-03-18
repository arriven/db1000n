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
		var packetConfig IPV4PacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}
		return buildIPV4Packet(packetConfig), nil
	default:
		return nil, fmt.Errorf("unsupported network layer type %s", c.Type)
	}
}

// IPV4PacketConfig describes ip layer configuration
type IPV4PacketConfig struct {
	SrcIP string `mapstructure:"src_ip"`
	DstIP string `mapstructure:"dst_ip"`
}

// buildIPV4Packet generates a layers.IPv4 and returns it with source IP address and destination IP address
func buildIPV4Packet(c IPV4PacketConfig) *layers.IPv4 {
	const ipv4 = 4

	return &layers.IPv4{
		SrcIP:    net.ParseIP(c.SrcIP).To4(),
		DstIP:    net.ParseIP(c.DstIP).To4(),
		Version:  ipv4,
		Protocol: layers.IPProtocolTCP,
	}
}
