// MIT License

// Copyright (c) [2022] [Arriven (https://github.com/Arriven)]

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

// Package packetgen [allows sending customized tcp/udp traffic. Inspired by https://github.com/bilalcaliskan/syn-flood]
package packetgen

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

// NetworkConfig describes which network to use when sending packets
type NetworkConfig struct {
	Name    string
	Address string
}

// PacketConfig stores full packet configuration
type PacketConfig struct {
	Ethernet EthernetPacketConfig
	IP       IPPacketConfig
	TCP      *TCPPacketConfig
	UDP      *UDPPacketConfig
	Payload  string
}

// BuildPacket is used to build and serialize packet based on the provided configuration
func BuildPacket(c PacketConfig) (payloadBuf gopacket.SerializeBuffer, ipHeader *ipv4.Header, err error) {
	var (
		udpPacket *layers.UDP
		tcpPacket *layers.TCP
	)
	ipPacket := buildIPPacket(c.IP)
	if c.UDP != nil {
		udpPacket = buildUDPPacket(*c.UDP)
		if err = udpPacket.SetNetworkLayerForChecksum(ipPacket); err != nil {
			return nil, nil, err
		}
	} else if c.TCP != nil {
		tcpPacket = buildTCPPacket(*c.TCP)
		if err = tcpPacket.SetNetworkLayerForChecksum(ipPacket); err != nil {
			return nil, nil, err
		}
	}

	// Serialize.  Note:  we only serialize the TCP layer, because the
	// socket we get with net.ListenPacket wraps our data in IPv4 packets
	// already.  We do still need the IP layer to compute checksums
	// correctly, though.
	ipHeaderBuf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if err = ipPacket.SerializeTo(ipHeaderBuf, opts); err != nil {
		return nil, nil, err
	}

	if ipHeader, err = ipv4.ParseHeader(ipHeaderBuf.Bytes()); err != nil {
		return nil, nil, err
	}

	ethernetLayer := buildEthernetPacket(c.Ethernet)
	payloadBuf = gopacket.NewSerializeBuffer()
	pyl := gopacket.Payload(c.Payload)

	if udpPacket != nil {
		if err = gopacket.SerializeLayers(payloadBuf, opts, ethernetLayer, udpPacket, pyl); err != nil {
			return nil, nil, err
		}
	} else if tcpPacket != nil {
		if err = gopacket.SerializeLayers(payloadBuf, opts, ethernetLayer, tcpPacket, pyl); err != nil {
			return nil, nil, err
		}
	}

	return payloadBuf, ipHeader, nil
}

// OpenRawConnection opens a raw ip network connection based on the provided config
func OpenRawConnection(c NetworkConfig) (*ipv4.RawConn, error) {
	var (
		packetConn net.PacketConn
		err        error
	)
	if packetConn, err = net.ListenPacket(c.Name, c.Address); err != nil {
		return nil, err
	}

	return ipv4.NewRawConn(packetConn)
}

// SendPacket is used to generate and send the packet over the network
func SendPacket(c PacketConfig, rawConn *ipv4.RawConn, destinationHost string, destinationPort int) (int, error) {
	var (
		payloadBuf gopacket.SerializeBuffer
		ipHeader   *ipv4.Header
		err        error
	)
	payloadBuf, ipHeader, err = BuildPacket(c)
	if err != nil {
		return 0, err
	}

	if err = rawConn.WriteTo(ipHeader, payloadBuf.Bytes(), nil); err != nil {
		return 0, err
	}
	return len(payloadBuf.Bytes()), nil
}

// IPPacketConfig describes ip layer configuration
type IPPacketConfig struct {
	SrcIP string `mapstructure:"src_ip"`
	DstIP string `mapstructure:"dst_ip"`
}

// buildIpPacket generates a layers.IPv4 and returns it with source IP address and destination IP address
func buildIPPacket(c IPPacketConfig) *layers.IPv4 {
	return &layers.IPv4{
		SrcIP:    net.ParseIP(c.SrcIP).To4(),
		DstIP:    net.ParseIP(c.DstIP).To4(),
		Version:  4,
		Protocol: layers.IPProtocolTCP,
	}
}

// UDPPacketConfig describes udp layer configuration
type UDPPacketConfig struct {
	SrcPort int `mapstructure:"src_port,string"`
	DstPort int `mapstructure:"dst_port,string"`
}

func buildUDPPacket(c UDPPacketConfig) *layers.UDP {
	return &layers.UDP{
		SrcPort: layers.UDPPort(c.SrcPort),
		DstPort: layers.UDPPort(c.DstPort),
	}
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
func buildTCPPacket(c TCPPacketConfig) *layers.TCP {
	return &layers.TCP{
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
