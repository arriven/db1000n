package packetgen

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/Arriven/db1000n/src/utils"
)

func BuildTransportLayer(c LayerConfig, network gopacket.NetworkLayer) (gopacket.TransportLayer, error) {
	switch c.Type {
	case "":
		return nil, nil
	case "tcp":
		var packetConfig TCPPacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return buildTCPPacket(packetConfig, network), nil
	case "udp":
		var packetConfig UDPPacketConfig
		if err := utils.Decode(c.Data, &packetConfig); err != nil {
			return nil, err
		}

		return buildUDPPacket(packetConfig, network), nil
	default:
		return nil, fmt.Errorf("unsupported link layer type %s", c.Type)
	}
}

// UDPPacketConfig describes udp layer configuration
type UDPPacketConfig struct {
	SrcPort int `mapstructure:"src_port,string"`
	DstPort int `mapstructure:"dst_port,string"`
}

func buildUDPPacket(c UDPPacketConfig, network gopacket.NetworkLayer) *layers.UDP {
	result := &layers.UDP{
		SrcPort: layers.UDPPort(c.SrcPort),
		DstPort: layers.UDPPort(c.DstPort),
	}
	if err := result.SetNetworkLayerForChecksum(network); err != nil {
		return nil
	}

	return result
}

// TCPFlagsConfig stores flags to be set on tcp layer
type TCPFlagsConfig struct {
	SYN bool
	ACK bool
	FIN bool
	RST bool
	PSH bool
	URG bool
	ECE bool
	CWR bool
	NS  bool
}

// TCPPacketConfig describes tcp layer configuration
type TCPPacketConfig struct {
	SrcPort int `mapstructure:"src_port,string"`
	DstPort int `mapstructure:"dst_port,string"`
	Seq     uint32
	Ack     uint32
	Window  uint16
	Urgent  uint16
	Flags   TCPFlagsConfig
}

// buildTCPPacket generates a layers.TCP and returns it with source port and destination port
func buildTCPPacket(c TCPPacketConfig, network gopacket.NetworkLayer) *layers.TCP {
	result := &layers.TCP{
		SrcPort: layers.TCPPort(c.SrcPort),
		DstPort: layers.TCPPort(c.DstPort),
		Window:  c.Window,
		Urgent:  c.Urgent,
		Seq:     c.Seq,
		Ack:     c.Ack,
		SYN:     c.Flags.SYN,
		ACK:     c.Flags.ACK,
		FIN:     c.Flags.FIN,
		RST:     c.Flags.RST,
		PSH:     c.Flags.PSH,
		URG:     c.Flags.URG,
		ECE:     c.Flags.ECE,
		CWR:     c.Flags.CWR,
		NS:      c.Flags.NS,
	}
	if err := result.SetNetworkLayerForChecksum(network); err != nil {
		return nil
	}

	return result
}
