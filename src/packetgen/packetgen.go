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

package packetgen

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

type PacketConfig struct {
	Ethernet EthernetPacketConfig
	IP       IPPacketConfig
	TCP      *TCPPacketConfig
	UDP      *UDPPacketConfig
	Payload  string
}

func SendPacket(c PacketConfig, destinationHost string, destinationPort int) (int, error) {
	var (
		ipHeader   *ipv4.Header
		packetConn net.PacketConn
		rawConn    *ipv4.RawConn
		udpPacket  *layers.UDP
		tcpPacket  *layers.TCP
		err        error
	)
	destinationHost, err = resolveHost(destinationHost)
	if err != nil {
		return 0, err
	}

	ipPacket := buildIpPacket(c.IP)
	if c.UDP != nil {
		udpPacket = buildUdpPacket(*c.UDP)
		if err = udpPacket.SetNetworkLayerForChecksum(ipPacket); err != nil {
			return 0, err
		}
	} else if c.TCP != nil {
		tcpPacket = buildTcpPacket(*c.TCP)
		if err = tcpPacket.SetNetworkLayerForChecksum(ipPacket); err != nil {
			return 0, err
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
		return 0, err
	}

	if ipHeader, err = ipv4.ParseHeader(ipHeaderBuf.Bytes()); err != nil {
		return 0, err
	}

	ethernetLayer := buildEthernetPacket(c.Ethernet)
	payloadBuf := gopacket.NewSerializeBuffer()
	pyl := gopacket.Payload(c.Payload)

	if udpPacket != nil {
		if err = gopacket.SerializeLayers(payloadBuf, opts, ethernetLayer, udpPacket, pyl); err != nil {
			return 0, err
		}

		// XXX send packet
		if packetConn, err = net.ListenPacket("ip4:udp", "0.0.0.0"); err != nil {
			return 0, err
		}
	} else if tcpPacket != nil {
		if err = gopacket.SerializeLayers(payloadBuf, opts, ethernetLayer, tcpPacket, pyl); err != nil {
			return 0, err
		}

		// XXX send packet
		if packetConn, err = net.ListenPacket("ip4:tcp", "0.0.0.0"); err != nil {
			return 0, err
		}
	}

	if rawConn, err = ipv4.NewRawConn(packetConn); err != nil {
		return 0, err
	}

	if err = rawConn.WriteTo(ipHeader, payloadBuf.Bytes(), nil); err != nil {
		return 0, err
	}
	return len(c.Payload), nil
}

type IPPacketConfig struct {
	SrcIP string `json:"src_ip"`
	DstIP string `json:"dst_ip"`
}

// buildIpPacket generates a layers.IPv4 and returns it with source IP address and destination IP address
func buildIpPacket(c IPPacketConfig) *layers.IPv4 {
	return &layers.IPv4{
		SrcIP:    net.ParseIP(c.SrcIP).To4(),
		DstIP:    net.ParseIP(c.DstIP).To4(),
		Version:  4,
		Protocol: layers.IPProtocolTCP,
	}
}

type UDPPacketConfig struct {
	SrcPort int `json:"src_port,string"`
	DstPort int `json:"dst_port,string"`
}

func buildUdpPacket(c UDPPacketConfig) *layers.UDP {
	return &layers.UDP{
		SrcPort: layers.UDPPort(c.SrcPort),
		DstPort: layers.UDPPort(c.DstPort),
	}
}

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

type TCPPacketConfig struct {
	SrcPort int `json:"src_port,string"`
	DstPort int `json:"dst_port,string"`
	Seq     uint32
	Ack     uint32
	Window  uint16
	Urgent  uint16
	Flags   TCPFlagsConfig
}

// buildTcpPacket generates a layers.TCP and returns it with source port and destination port
func buildTcpPacket(c TCPPacketConfig) *layers.TCP {
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

type EthernetPacketConfig struct {
	SrcMAC string `json:"src_mac"`
	DstMAC string `json:"dst_mac"`
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
