package packetgen

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/Arriven/db1000n/src/utils"
)

func BuildLinkLayer(c LayerConfig) (gopacket.LinkLayer, error) {
	switch c.Type {
	case "":
		return nil, nil
	case "ethernet":
		var packetConfig EthernetPacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return buildEthernetPacket(packetConfig), nil
	default:
		return nil, fmt.Errorf("unsupported link layer type %s", c.Type)
	}
}

// EthernetPacketConfig describes ethernet layer configuration
type EthernetPacketConfig struct {
	SrcMAC string `mapstructure:"src_mac"`
	DstMAC string `mapstructure:"dst_mac"`
}

// buildEthernetPacket generates an layers.Ethernet and returns it with source MAC address and destination MAC address
func buildEthernetPacket(c EthernetPacketConfig) *layers.Ethernet {
	srcMac := net.HardwareAddr(c.SrcMAC)
	dstMac := net.HardwareAddr(c.DstMAC)

	return &layers.Ethernet{
		SrcMAC: net.HardwareAddr{srcMac[0], srcMac[1], srcMac[2], srcMac[3], srcMac[4], srcMac[5]},
		DstMAC: net.HardwareAddr{dstMac[0], dstMac[1], dstMac[2], dstMac[3], dstMac[4], dstMac[5]},
	}
}
